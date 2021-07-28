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
  // Alphabetize lists by URL to show duplicates
  // get id endpoints
  getAvailableArc: (id) => `/api/available-articles/${id}`,
  getScheduledArticle: (id) => `/api/scheduled-articles/${id}`,
  getPage: (id) => `/api/page/${id}`,
  // list page points
  listAnyArc: (page = "0") => `/api/list-any-arc/${page}`,
  listAvailableArc: (page = "0") => `/api/list-available/${page}`,
  listNewsletterPages: (page = "0") => `/api/newsletter-pages/${page}`,
  // GET and POST listed as two endpoints
  listAllSeries: `/api/all-series`,
  listAllTopics: `/api/all-topics`,
  postAuthorizedDomain: `/api/authorized-domains`,
  listAuthorizedDomains: `/api/authorized-domains`,
  postAuthorizedEmailAddress: `/api/authorized-addresses`,
  listAuthorizedEmailAddresses: `/api/authorized-addresses`,
  saveArcArticle: `/api/available-articles`,
  createSignedUpload: `/api/create-signed-upload`,
  getEditorsPicks: `/api/editors-picks`,
  saveEditorsPicks: `/api/editors-picks`,
  createFile: `/api/files-create`,
  listFiles: `/api/files-list`,
  updateFile: `/api/files-update`,
  updateImage: `/api/image-update`,
  listImages: `/api/images`,
  listRefreshArc: `/api/list-arc-refresh`,
  getSignupURL: `/api/mailchimp-signup-url`,
  sendMessage: `/api/message`,
  postPage: `/api/page`,
  saveArticle: `/api/scheduled-articles`,
  listSpotlightPAArticles: `/api/spotlightpa-articles`,
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
    async uploadImage(body) {
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
    async uploadFile(body) {
      let [data, err] = await tryTo(
        post(endpoints.createFile, { filename: body.name, mimeType: body.type })
      );
      if (err) {
        return ["", err];
      }
      let {
        "signed-url": signedURL,
        "file-url": fileURL,
        "cache-control": cacheControl,
        disposition,
      } = data;
      let opts = {
        method: "PUT",
        body,
        headers: {
          "Content-Disposition": disposition,
          "Cache-Control": cacheControl,
        },
      };
      let rsp;
      [rsp, err] = await tryTo(fetch(signedURL, opts));
      if (err ?? !rsp.ok) {
        return ["", err ?? responseError(rsp)];
      }
      [, err] = await actions.updateFile(fileURL);
      if (err) {
        return ["", err];
      }
      return [fileURL, null];
    },
    async updateFile(url, { description = null } = {}) {
      let file = {
        url,
        description,
        set_description: description !== null,
      };
      return await tryTo(post(endpoints.updateFile, file));
    },
  };
  let idGetActions = [
    // does not include proxy images…
    "getAvailableArc",
    "getScheduledArticle",
    "getPage",
    "listAnyArc",
    "listAvailableArc",
    "listNewsletterPages",
  ];
  for (let action of idGetActions) {
    let endpointFn = endpoints[action];
    actions[action] = (id) => tryTo(request(endpointFn(id)));
  }
  let simpleGetActions = [
    "getEditorsPicks",
    "getSignupURL",
    "listAllSeries",
    "listAllTopics",
    "listAuthorizedDomains",
    "listAuthorizedEmailAddresses",
    "listFiles",
    "listImages",
    "listRefreshArc",
    "listSpotlightPAArticles",
  ];
  for (let action of simpleGetActions) {
    let endpoint = endpoints[action];
    actions[action] = () => tryTo(request(endpoint));
  }
  let simplePostActions = [
    "postAuthorizedDomain",
    "postAuthorizedEmailAddress",
    "postPage",
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
