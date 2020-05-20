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
  addAuthorizedDomain: `/api/authorized-domains`,
  listAuthorizedDomains: `/api/authorized-domains`,
  listAvailable: `/api/available-articles`,
  postAvailable: `/api/available-articles`,
  createSignedUpload: `/api/create-signed-upload`,
  getEditorsPicks: `/api/editors-picks`,
  saveEditorsPicks: `/api/editors-picks`,
  updateImage: `/api/image-update`,
  listImages: `/api/images`,
  listRefreshArc: `/api/list-arc-refresh`,
  getSignupURL: `/api/mailchimp-signup-url`,
  sendMessage: `/api/message`,
  saveArticle: `/api/scheduled-articles`,
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
    async article(id) {
      return await tryTo(request(endpoints.scheduledArticle(id)));
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
  };
  let simpleGetActions = [
    "getEditorsPicks",
    "getSignupURL",
    "listAuthorizedDomains",
    "listAvailable",
    "listImages",
    "listRefreshArc",
    "listSpotlightPAArticles",
    "upcoming",
  ];
  for (let action of simpleGetActions) {
    actions[action] = () => tryTo(request(endpoints[action]));
  }
  let simplePostActions = [
    "addAuthorizedDomain",
    "postAvailable",
    "saveArticle",
    "saveEditorsPicks",
    "sendMessage",
  ];
  for (let action of simplePostActions) {
    actions[action] = (obj) => tryTo(post(endpoints[action], obj));
  }

  return actions;
}
