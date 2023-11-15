import { reactive, computed, toRefs } from "vue";

import netlifyIdentity from "netlify-identity-widget";

function makeAuth() {
  let authState = reactive({
    user: null,
  });

  netlifyIdentity.on("init", async (user) => {
    authState.user = user;
    try {
      await user.jwt();
    } catch (err) {
      await methods.logout();
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
    // eslint-disable-next-line no-console
    console.warn(err);
    authState.user = null;
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
    async logout() {
      authState.user = null;
      try {
        await netlifyIdentity.logout();
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn(e);
        netlifyIdentity.store.user = null;
      }
    },
    async headers() {
      if (!authState.user) {
        return null;
      }
      let token;
      try {
        token = await authState.user.jwt();
      } catch (e) {
        await methods.logout();
        return null;
      }
      return {
        Authorization: `Bearer ${token}`,
      };
    },
  };

  let APIUrl = window.location.hostname.match(/localhost|\.ts\.net/)
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

let $auth;

export function useAuth() {
  if (!$auth) {
    $auth = makeAuth();
  }
  return $auth;
}
