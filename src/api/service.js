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
  healthcheck: `/api/healthcheck`,
  userInfo: `/api/user-info`,
  listAvailable: `/api/available-articles`,
  getAvailable: (id) => `/api/available-articles/${id}`,
  postAvailable: `/api/available-articles`,
  upcoming: `/api/upcoming-articles`,
  listRefreshArc: `/api/list-arc-refresh`,
  getMessage: (id) => `/api/message/${id}`,
  sendMessage: `/api/message`,
  scheduledArticle: (id) => `/api/scheduled-articles/${id}`,
  scheduleArticle: `/api/scheduled-articles`,
  createSignedUpload: `/api/create-signed-upload`,
  updateImage: `/api/image-update`,
  getSignupURL: `/api/mailchimp-signup-url`,
  authorizedDomains: `/api/authorized-domains`,
  listSpotlightPAArticles: `/api/spotlightpa-articles`,
  listImages: `/api/images`,
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

  let requestBuffer = {};
  async function bufferRequest(key, cb) {
    if (requestBuffer[key]) {
      // eslint-disable-next-line no-console
      console.warning("buffering request", key);
      return await requestBuffer[key];
    }
    let promise = cb();
    requestBuffer[key] = promise;
    let r = await promise;
    requestBuffer[key] = null;
    return r;
  }

  let actions = {
    async userInfo() {
      return await bufferRequest("userInfo", () =>
        tryTo(request(endpoints.userInfo))
      );
    },
    async listAvailable() {
      return await tryTo(request(endpoints.listAvailable));
    },
    hasAuthAvailable() {
      return $auth.isEditor;
    },
    async getAvailable(id) {
      return await tryTo(request(endpoints.getAvailable(id)));
    },
    async upcoming() {
      return await tryTo(request(endpoints.upcoming));
    },
    hasAuthUpcoming() {
      return $auth.isSpotlightPAUser;
    },
    async postAvailable(obj) {
      return await tryTo(post(endpoints.postAvailable, obj));
    },
    async article(id) {
      return await tryTo(request(endpoints.scheduledArticle(id)));
    },
    hasAuthArticle() {
      return $auth.isSpotlightPAUser;
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
    async getSignupURL() {
      return await tryTo(request(endpoints.getSignupURL));
    },
    async listAuthorizedDomains() {
      return await tryTo(request(endpoints.authorizedDomains));
    },
    async addAuthorizedDomain(domain) {
      return await tryTo(post(endpoints.authorizedDomains, { domain }));
    },
    async listSpotlightPAArticles() {
      return await tryTo(request(endpoints.listSpotlightPAArticles));
    },
    async listRefreshArc() {
      return await tryTo(request(endpoints.listRefreshArc));
    },
    async listImages() {
      return await tryTo(request(endpoints.listImages));
    },
  };
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
