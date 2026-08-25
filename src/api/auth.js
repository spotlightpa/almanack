import { reactive, computed, toRefs } from "vue";

import netlifyIdentity from "netlify-identity-widget";

const DEV_AUTH_KEY = "almanack_dev_auth";

function loadDevUser() {
  try {
    return JSON.parse(localStorage.getItem(DEV_AUTH_KEY)) ?? null;
  } catch {
    return null;
  }
}

function saveDevUser(user) {
  if (user) {
    localStorage.setItem(DEV_AUTH_KEY, JSON.stringify(user));
  } else {
    localStorage.removeItem(DEV_AUTH_KEY);
  }
}

function makeDevAuth() {
  const authState = reactive({
    user: loadDevUser(),
  });

  const isSignedIn = computed(() => !!authState.user);
  const roles = computed(() => authState.user?.roles ?? []);
  const fullName = computed(() => authState.user?.fullName ?? "");
  const email = computed(() => authState.user?.email ?? "");

  function hasRole(name) {
    return computed(() =>
      roles.value.some((role) => role === name || role === "admin")
    );
  }

  return {
    ...toRefs(authState),
    isSignedIn,
    roles,
    fullName,
    email,
    isEditor: hasRole("editor"),
    isSpotlightPAUser: hasRole("Spotlight PA"),
    isArcUser: hasRole("arc user"),
    signup() {},
    login() {},
    async logout() {
      authState.user = null;
      saveDevUser(null);
    },
    async headers() {
      if (!authState.user) return null;
      return { Authorization: "Bearer dev-fake-token" };
    },
    setDevUser({ email, fullName, roles }) {
      let user = { email, fullName, roles };
      authState.user = user;
      saveDevUser(user);
    },
  };
}

function makeAuth() {
  const authState = reactive({
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
    console.warn(err);
    authState.user = null;
  });

  const token = computed(() => authState.user?.token?.access_token ?? null);
  const isSignedIn = computed(() => !!token.value);
  const roles = computed(() => authState.user?.app_metadata?.roles ?? []);
  const fullName = computed(
    () => authState.user?.user_metadata?.full_name ?? ""
  );
  const email = computed(() => authState.user?.email ?? "");

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
    $auth = import.meta.env.MODE !== "production" ? makeDevAuth() : makeAuth();
  }
  return $auth;
}
