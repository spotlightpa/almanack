import { useAuth } from "./auth.js";

const tryTo = (promise) =>
  promise
    // Wrap data/errors
    .then((data) => [data, null])
    .catch((error) => [null, error]);

const responseError = (rsp) => {
  if (rsp.ok) {
    return;
  }
  let msg = `${rsp.status} ${rsp.statusText}`;
  let err = new Error("Unexpected response from server: " + msg);
  err.name = msg;
  return err;
};

const endpoints = {
  getAvailableArc: (id) => `/api/available-articles/${id}`,
  getScheduledArticle: (id) => `/api/scheduled-articles/${id}`,
  // Alphabetized by URL to show duplicates
  // GET and POST listed as two endpoints
  addAuthorizedDomain: `/api/authorized-domains`,
  listAuthorizedDomains: `/api/authorized-domains`,
  listAvailableArc: `/api/available-articles`,
  saveArcArticle: `/api/available-articles`,
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
  listAnyArc: `/api/upcoming-articles`,
};

function makeClient($auth) {
  async function request(url, options = {}) {
    let headers = await $auth.headers();
    if (!headers) {
      let err = new Error("Please log in again.");
      err.name = "Login Error";
      throw err;
    }
    if (options.headers) {
      options.headers = { ...headers, ...options.headers };
    }
    let defaultOpts = {
      headers,
    };
    options = { ...defaultOpts, ...options };
    let resp = await fetch(url, options);
    let err = responseError(resp);
    if (err) throw err;

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
  let idGetActions = [
    // does not include proxy images…
    "getAvailableArc",
    "getScheduledArticle",
  ];
  for (let action of idGetActions) {
    let endpointFn = endpoints[action];
    actions[action] = (id) => tryTo(request(endpointFn(id)));
  }
  let simpleGetActions = [
    "getEditorsPicks",
    "getSignupURL",
    "listAnyArc",
    "listAuthorizedDomains",
    "listAvailableArc",
    "listImages",
    "listRefreshArc",
    "listSpotlightPAArticles",
  ];
  for (let action of simpleGetActions) {
    let endpoint = endpoints[action];
    actions[action] = () => tryTo(request(endpoint));
  }
  let simplePostActions = [
    "addAuthorizedDomain",
    "saveArcArticle",
    "saveArticle",
    "saveEditorsPicks",
    "sendMessage",
  ];
  for (let action of simplePostActions) {
    let endpoint = endpoints[action];
    actions[action] = (obj) => tryTo(post(endpoint, obj));
  }

  return actions;
}

let $client;

export function useClient() {
  if (!$client) {
    $client = makeClient(useAuth());
  }
  return $client;
}
