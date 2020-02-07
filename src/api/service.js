const tryTo = promise =>
  promise
    // Wrap data/errors
    .then(data => [data, null])
    .catch(error => [null, error]);

export const endpoints = {
  userInfo: `/api/user-info`,
  upcoming: `/api/upcoming`,
  getArticle(id) {
    return `/api/articles/${id}`;
  },
  getSignedUpload: `/api/get-signed-upload`,
};

export function makeService($auth) {
  async function request(url, options = {}) {
    let headers = await $auth.headers();
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
    async upcoming() {
      return await bufferRequest("upcoming", () =>
        tryTo(request(endpoints.upcoming))
      );
    },
    hasAuthUpcoming() {
      return $auth.isEditor;
    },
    async article(id) {
      return await tryTo(request(endpoints.getArticle(id)));
    },
    hasAuthArticle() {
      return $auth.isSpotlightPAUser;
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
