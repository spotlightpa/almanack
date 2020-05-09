import { reactive, computed, toRefs } from "@vue/composition-api";

import netlifyIdentity from "netlify-identity-widget";

export function makeAuth() {
  let authState = reactive({
    user: null,
    error: null,
  });

  netlifyIdentity.on("init", async (user) => {
    authState.user = user;
    try {
      await user.jwt();
    } catch (err) {
      authState.error = err;
      methods.logout();
    }
  });
  netlifyIdentity.on("login", (user) => {
    authState.user = user;
    netlifyIdentity.close();
  });
  netlifyIdentity.on("logout", () => {
    authState.user = null;
  });
  netlifyIdentity.on("error", (err) => {
    authState.error = err;
  });

  let token = computed(() => authState.user?.token?.access_token ?? null);
  let isSignedIn = computed(() => !!token.value);
  let roles = computed(() => authState.user?.app_metadata?.roles ?? []);
  let fullName = computed(() => authState.user?.user_metadata?.full_name ?? "");
  let email = computed(() => authState.user?.email ?? "");

  function hasRole(name) {
    return computed(() => {
      return roles.value.some((role) => role === name || role === "admin");
    });
  }

  let methods = {
    signup() {
      netlifyIdentity.open("signup");
    },
    login() {
      netlifyIdentity.open("login");
    },
    logout() {
      authState.user = null;
      try {
        netlifyIdentity.logout();
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn(e);
      }
    },
    async headers() {
      let token;
      try {
        token = await authState.user.jwt();
      } catch (e) {
        methods.logout();
        throw e;
      }
      return {
        Authorization: `Bearer ${token}`,
      };
    },
  };

  let APIUrl = window.location.hostname.match(/localhost/)
    ? "https://almanack.data.spotlightpa.org/.netlify/identity"
    : null;
  netlifyIdentity.init({ logo: false, APIUrl });

  return {
    ...toRefs(authState),

    isSignedIn,
    roles,
    fullName,
    email,

    isEditor: hasRole("editor"),
    isSpotlightPAUser: hasRole("Spotlight PA"),
    isArcUser: hasRole("arc user"),

    ...methods,
  };
}
