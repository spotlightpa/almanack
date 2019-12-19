const tryTo = promise =>
  promise
    // Wrap data/errors
    .then(data => [data, null])
    .catch(error => [null, error]);

export const endpoints = {
  userInfo: `/api/user-info`,
  upcoming: `/api/upcoming`,
};

export function makeService($auth) {
  async function request(url, options = {}) {
    let headers = await $auth.headers();
    let defaultOpts = {
      headers,
    };
    options = { ...defaultOpts, options };
    let resp = await fetch(url, options);
    if (!resp.ok) {
      throw new Error(
        `Unexpected response from server (status ${resp.status})`
      );
    }
    return await resp.json();
  }

  return {
    async userInfo() {
      return await tryTo(request(endpoints.userInfo));
    },
    async upcoming() {
      return await tryTo(request(endpoints.upcoming));
    },
  };
}
