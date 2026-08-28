import { ref, computed } from "vue";
import netlifyIdentity from "netlify-identity-widget";
import type { NetlifyUser } from "netlify-identity-widget";

const DEV_AUTH_KEY = "almanack_dev_auth";

interface DevUser {
  email: string;
  fullName: string;
  roles: string[];
}

function loadDevUser(): DevUser | null {
  try {
    return (
      (JSON.parse(
        localStorage.getItem(DEV_AUTH_KEY) ?? "null"
      ) as DevUser | null) ?? null
    );
  } catch {
    return null;
  }
}

function saveDevUser(user: DevUser | null) {
  if (user) {
    localStorage.setItem(DEV_AUTH_KEY, JSON.stringify(user));
  } else {
    localStorage.removeItem(DEV_AUTH_KEY);
  }
}

function makeDevAuth() {
  const user = ref<DevUser | null>(loadDevUser());

  const isSignedIn = computed(() => !!user.value);
  const roles = computed(() => user.value?.roles ?? []);
  const fullName = computed(() => user.value?.fullName ?? "");
  const email = computed(() => user.value?.email ?? "");

  function hasRole(name: string) {
    return computed(() =>
      roles.value.some((role) => role === name || role === "admin")
    );
  }

  return {
    user,
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
      user.value = null;
      saveDevUser(null);
    },
    async headers(): Promise<Record<string, string> | null> {
      if (!user.value) return null;
      return { Authorization: "Bearer dev-fake-token" };
    },
    setDevUser({ email, fullName, roles }: DevUser) {
      user.value = { email, fullName, roles };
      saveDevUser(user.value);
    },
  };
}

function makeAuth() {
  const user = ref<NetlifyUser | null>(null);

  const token = computed(() => user.value?.token?.access_token ?? null);
  const isSignedIn = computed(() => !!token.value);
  const roles = computed(() => user.value?.app_metadata?.roles ?? []);
  const fullName = computed(() => user.value?.user_metadata?.full_name ?? "");
  const email = computed(() => user.value?.email ?? "");

  function hasRole(name: string) {
    return computed(() =>
      roles.value.some((role) => role === name || role === "admin")
    );
  }

  const methods = {
    signup() {
      netlifyIdentity.open("signup");
    },
    login() {
      netlifyIdentity.open("login");
    },
    async logout() {
      user.value = null;
      try {
        await netlifyIdentity.logout();
      } catch (e) {
        console.warn(e);
        netlifyIdentity.store.user = null;
      }
    },
    async headers(): Promise<Record<string, string> | null> {
      if (!user.value) {
        return null;
      }
      let token: string;
      try {
        token = await user.value.jwt();
      } catch (e) {
        await methods.logout();
        return null;
      }
      return {
        Authorization: `Bearer ${token}`,
      };
    },
  };

  netlifyIdentity.on("init", async (u) => {
    user.value = u;
    try {
      await u?.jwt();
    } catch {
      await methods.logout();
    }
  });
  netlifyIdentity.on("login", (u) => {
    user.value = u;
    netlifyIdentity.close();
  });
  netlifyIdentity.on("logout", () => {
    user.value = null;
  });
  netlifyIdentity.on("error", (err) => {
    console.warn(err);
    user.value = null;
  });

  const APIUrl = window.location.hostname.match(/localhost|\.ts\.net/)
    ? "https://almanack.data.spotlightpa.org/.netlify/identity"
    : null;
  netlifyIdentity.init({ logo: false, APIUrl });

  return {
    user,
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

let $auth: ReturnType<typeof makeDevAuth> | ReturnType<typeof makeAuth> | null =
  null;

export function useAuth() {
  if (!$auth) {
    $auth = import.meta.env.MODE !== "production" ? makeDevAuth() : makeAuth();
  }
  return $auth;
}
