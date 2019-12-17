import { reactive, computed, toRefs } from "@vue/composition-api";

import netlifyIdentity from "netlify-identity-widget";

function makeAuth() {
  let data = reactive({
    user: null,
    error: null,
  });

  netlifyIdentity.on("init", async user => {
    data.user = user;
    try {
      await user.jwt();
    } catch (err) {
      data.error = err;
      logout();
    }
  });
  netlifyIdentity.on("login", user => {
    data.user = user;
    netlifyIdentity.close();
  });
  netlifyIdentity.on("logout", () => {
    data.user = null;
  });
  netlifyIdentity.on("error", err => {
    data.error = err;
  });
  netlifyIdentity.init({ logo: false });

  let token = computed(
    () => (data.user && data.user.token && data.user.token.access_token) || null
  );
  let isSignedIn = computed(() => !!token.value);
  let roles = computed(
    () =>
      (isSignedIn.value &&
        data.user.app_metadata &&
        data.user.app_metadata.roles) ||
      []
  );

  let fullName = computed(
    () =>
      (data.user &&
        data.user.user_metadata &&
        data.user.user_metadata.full_name) ||
      ""
  );

  function logout() {
    data.user = null;
    netlifyIdentity.logout();
  }

  return {
    ...toRefs(data),

    isSignedIn,
    roles,
    fullName,
    logout,

    signup() {
      netlifyIdentity.open("signup");
    },
    login() {
      netlifyIdentity.open("login");
    },
    async headers() {
      let token;
      try {
        token = await data.user.jwt(true);
      } catch (e) {
        logout();
        throw e;
      }
      return {
        Authorization: `Bearer ${token}`,
      };
    },
    hasRole(name) {
      return roles.value.some(role => role === name || role === "admin");
    },
  };
}

let $auth;

export function useAuth() {
  if (!$auth) {
    $auth = makeAuth();
  }
  return $auth;
}
