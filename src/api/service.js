import { reactive, toRefs } from "@vue/composition-api";

const tryTo = promise =>
  promise
    // Wrap data/errors
    .then(data => [data, null])
    .catch(error => [null, error]);

export const endpoints = {
  healthcheck: `/api/healthcheck`,
  userInfo: `/api/user-info`,
  listAvailable: `/api/available-articles`,
  available: id => `/api/available-articles/${id}`,
  makePlanned: id => `/api/available-articles/${id}/planned`,
  makeAvailable: id => `/api/available-articles/${id}/available`,
  upcoming: `/api/upcoming-articles`,
  getMessage: id => `/api/message/${id}`,
  sendMessage: `/api/message`,
  scheduledArticle: id => `/api/scheduled-articles/${id}`,
  scheduleArticle: `/api/scheduled-articles`,
  getSignedUpload: `/api/get-signed-upload`,
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

  return {
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
    async available(id) {
      return await tryTo(request(endpoints.available(id)));
    },
    async upcoming() {
      return await tryTo(request(endpoints.upcoming));
    },
    hasAuthUpcoming() {
      return $auth.isSpotlightPAUser;
    },
    async makePlanned(id) {
      return await tryTo(post(endpoints.makePlanned(id), null));
    },
    async makeAvailable(id) {
      return await tryTo(post(endpoints.makeAvailable(id), null));
    },
    async removedArticle(id) {
      return await tryTo(
        request(endpoints.available(id), { method: "DELETE" })
      );
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
        request(endpoints.getSignedUpload, { method: "POST" })
      );
      if (err) {
        return ["", err];
      }
      let { "signed-url": signedURL, filename } = data;
      let rsp;
      [rsp, err] = await tryTo(fetch(signedURL, { method: "PUT", body }));
      if (err ?? !rsp.ok) {
        return [
          "",
          err ??
            new Error(`Unexpected response: ${rsp.status} ${rsp.statusText}`),
        ];
      }
      return [filename, null];
    },
  };
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
    async fetch({ arg = null, force = false } = {}) {
      if (!apiState.canLoad) {
        apiState.error = new Error("Insufficient permissions");
        apiState.error.name = "Unauthorized";
        return;
      }
      if (apiState.isLoading && !force) {
        return;
      }
      apiState.isLoading = true;
      [apiState.rawData, apiState.error] = await serviceCall(arg);
      apiState.isLoading = false;
      apiState.didLoad = true;
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
