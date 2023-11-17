import { computed, ref, effect } from "vue";

import GoTrue from "gotrue-js";

const auth = new GoTrue({
  APIUrl: "https://almanack.data.spotlightpa.org/.netlify/identity",
  setCookie: true,
});

const lastModified = ref(1);
const loginError = ref(null);

effect(() => {
  if (loginError.value) console.log("loginError", loginError.value);
});

async function init() {
  const errorRoute = /error=access_denied&error_description=403/;
  const typeTokenRoutes =
    /(confirmation|invite|recovery|email_change)_token=([^&]+)/;
  const accessTokenRoute = /access_token=/;

  let hash = document.location.hash.slice(1); // remove # at start
  if (!hash) {
    return;
  }

  let match = hash.match(errorRoute);
  if (match) {
    loginError.value = "Access denied";
    document.location.hash = "";
    return;
  }

  match = hash.match(typeTokenRoutes);
  if (match) {
    loginError.value = null;
    let [, type, token] = match;
    await auth.verify(type, token, true);
    lastModified.value++;
    document.location.hash = "";
    return;
  }

  match = hash.match(accessTokenRoute);
  if (match) {
    loginError.value = null;
    let searchParams = new URLSearchParams(hash);
    let params = {
      access_token: searchParams.get("access_token"),
      expires_in: searchParams.get("expires_in"),
      refresh_token: searchParams.get("refresh_token"),
      token_type: searchParams.get("token_type"),
    };
    await auth.createUser(params, true);
    lastModified.value++;
    document.location.hash = "";
  }
}

export const fullName = computed(
  () =>
    lastModified.value && (auth.currentUser()?.user_metadata?.full_name ?? "")
);

export async function signup({ fullName, email, password }) {
  loginError.value = null;
  try {
    await auth.signup(email, password, { full_name: fullName });
  } catch (err) {
    lastModified.value++;
    loginError.value = err;
    return;
  }
  await login(email, password);
}

export async function login(email, password) {
  loginError.value = null;
  try {
    await auth.login(email, password, true);
  } catch (err) {
    loginError.value = err;
  }
  lastModified.value++;
}

export async function logout() {
  loginError.value = null;
  await auth.currentUser().logout();
  lastModified.value++;
}

export async function headers() {
  let user = auth.currentUser();
  if (!user) {
    return null;
  }
  let token;
  try {
    token = await user.jwt();
  } catch (e) {
    await logout();
    return null;
  }
  return {
    Authorization: `Bearer ${token}`,
  };
}

init().catch((err) => {
  loginError.value = err;
  lastModified.value++;
});
