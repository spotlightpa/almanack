import { reactive, toRefs } from "@vue/composition-api";

const tryTo = (promise) =>
  promise
    // Wrap data/errors
    .then((data) => [data, null])
    .catch((error) => [null, error]);

const responseError = (rsp) => {
  if (rsp.ok) {
    return;
  }
  let err = new Error(`Unexpected response: ${rsp.status} ${rsp.statusText}`);
  err.name = "Unexpected Response";
  return err;
};

export const endpoints = {
  getAvailable: (id) => `/api/available-articles/${id}`,
  scheduledArticle: (id) => `/api/scheduled-articles/${id}`,
  authorizedDomains: `/api/authorized-domains`,
  listAvailable: `/api/available-articles`,
  postAvailable: `/api/available-articles`,
  createSignedUpload: `/api/create-signed-upload`,
  editorsPicks: `/api/editors-picks`,
  updateImage: `/api/image-update`,
  listImages: `/api/images`,
  listRefreshArc: `/api/list-arc-refresh`,
  getSignupURL: `/api/mailchimp-signup-url`,
  sendMessage: `/api/message`,
  scheduleArticle: `/api/scheduled-articles`,
  listSpotlightPAArticles: `/api/spotlightpa-articles`,
  upcoming: `/api/upcoming-articles`,
};

export function makeClient($auth) {
  async function request(url, options = {}) {
    let headers = await $auth.headers();
    if (options.headers) {
      options.headers = { ...headers, ...options.headers };
    }
    let defaultOpts = {
      headers,
    };
    options = { ...defaultOpts, ...options };
    let resp = await fetch(url, options);
    if (!resp.ok) {
      throw new Error(
        `Unexpected response from server (status ${resp.status})`
      );
    }
    return await resp.json();
  }

  function post(url, obj) {
    let body = JSON.stringify(obj);
    return request(url, {
      headers: { "Content-Type": "application/json" },
      method: "POST",
      body,
    });
  }

  let actions = {
    hasAuthAvailable() {
      return $auth.isEditor;
    },
    hasAuthUpcoming() {
      return $auth.isSpotlightPAUser;
    },
    hasAuthArticle() {
      return $auth.isSpotlightPAUser;
    },
    async getAvailable(id) {
      return await tryTo(request(endpoints.getAvailable(id)));
    },
    async postAvailable(obj) {
      return await tryTo(post(endpoints.postAvailable, obj));
    },
    async article(id) {
      return await tryTo(request(endpoints.scheduledArticle(id)));
    },
    async saveArticle(id, obj) {
      return await tryTo(post(endpoints.scheduleArticle, obj));
    },
    async sendMessage(obj) {
      return await tryTo(post(endpoints.sendMessage, obj));
    },
    async uploadFile(body) {
      let [data, err] = await tryTo(
        post(endpoints.createSignedUpload, { type: body.type })
      );
      if (err) {
        return ["", err];
      }
      let { "signed-url": signedURL, filename } = data;
      let rsp;
      [rsp, err] = await tryTo(fetch(signedURL, { method: "PUT", body }));
      if (err ?? !rsp.ok) {
        return ["", err ?? responseError(rsp)];
      }
      [, err] = await actions.updateImage(filename);
      if (err) {
        return ["", err];
      }
      return [filename, null];
    },
    async updateImage(path, { credit = "", description = "" } = {}) {
      let image = {
        path,
        credit,
        set_credit: !!credit,
        description,
        set_description: !!description,
      };
      return await tryTo(post(endpoints.updateImage, image));
    },
    async addAuthorizedDomain(domain) {
      return await tryTo(post(endpoints.authorizedDomains, { domain }));
    },
    async saveEditorsPicks(obj) {
      return await tryTo(post(endpoints.editorsPicks, obj));
    },
  };
  let simpleActions = [
    ["getEditorsPicks", "editorsPicks"],
    ["getSignupURL", "getSignupURL"],
    ["listAuthorizedDomains", "authorizedDomains"],
    ["listAvailable", "listAvailable"],
    ["listImages", "listImages"],
    ["listRefreshArc", "listRefreshArc"],
    ["listSpotlightPAArticles", "listSpotlightPAArticles"],
    ["upcoming", "upcoming"],
  ];
  for (let [name, endpoint] of simpleActions) {
    actions[name] = () => tryTo(request(endpoints[endpoint]));
  }
  return actions;
}

export function useService({ canLoad, serviceCall }) {
  const apiState = reactive({
    rawData: null,
    isLoading: false,
    error: null,
    didLoad: false,
    canLoad,
  });

  let methods = {
    async do(callback, { force = false } = {}) {
      if (!apiState.canLoad) {
        apiState.error = new Error("Insufficient permissions");
        apiState.error.name = "Unauthorized";
        return;
      }
      if (apiState.isLoading && !force) {
        return;
      }
      apiState.isLoading = true;
      [apiState.rawData, apiState.error] = await callback();
      apiState.isLoading = false;
      apiState.didLoad = true;
    },
    async fetch({ arg = null, force = false } = {}) {
      await methods.do(() => serviceCall(arg), { force });
    },
    async initLoad() {
      if (apiState.canLoad && !apiState.didLoad) {
        await methods.fetch();
      }
    },
  };

  return {
    ...toRefs(apiState),
    ...methods,
  };
}
