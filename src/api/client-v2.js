import { useAuth } from "./auth.js";

const tryTo = (promise) =>
  promise
    // Wrap data/errors
    .then((data) => [data, null])
    .catch((error) => [null, error]);

const responseError = async (rsp) => {
  if (rsp.ok) {
    return;
  }
  let details = {};
  try {
    details = (await rsp.json())?.details ?? {};
    // eslint-disable-next-line no-empty
  } catch (e) {}

  let msg = `${rsp.status} ${rsp.statusText}`;
  let err = new Error("Unexpected response from server: " + msg);
  err.name = msg;
  err.details = details;
  return err;
};

let $auth = null;

async function request(
  url,
  { params = null, headers = {}, options = {} } = {}
) {
  if (!$auth) {
    $auth = useAuth();
  }

  let authHeaders = await $auth.headers();
  if (!authHeaders) {
    let err = new Error("Please log in again.");
    err.name = "Login Error";
    throw err;
  }

  headers = { ...authHeaders, ...headers };

  if (params) {
    url += `?${new URLSearchParams(params)}`;
  }
  options = { headers, ...options };
  let resp = await fetch(url, options);
  let err = await responseError(resp);
  if (err) throw err;

  return await resp.json();
}

export function get(url, params) {
  return tryTo(request(url, { params }));
}

export function post(url, obj) {
  let body = JSON.stringify(obj);
  return tryTo(
    request(url, {
      headers: { "Content-Type": "application/json" },
      options: {
        method: "POST",
        body,
      },
    })
  );
}

// Alphabetize lists by URL to show duplicates
// GET and POST listed as two endpoints
export const listAllSeries = `/api/all-series`;
export const listAllTopics = `/api/all-topics`;
export const postAuthorizedDomain = `/api/authorized-domains`;
export const listAuthorizedDomains = `/api/authorized-domains`;
export const postAuthorizedEmailAddress = `/api/authorized-addresses`;
export const listAuthorizedEmailAddresses = `/api/authorized-addresses`;
export const createSignedUpload = `/api/create-signed-upload`;
export const postDonorWall = `/api/donor-wall`;
export const createFile = `/api/files-create`;
export const listFiles = `/api/files-list`;
export const updateFile = `/api/files-update`;
export const getGDocsDoc = `/api/gdocs-doc`;
export const postGDocsDoc = `/api/gdocs-doc`;
export const postImageUpdate = `/api/image-update`;
export const listImages = `/api/images`;
export const sendMessage = `/api/message`;
export const getPage = `/api/page`;
export const postPage = `/api/page`;
export const postPageCreate = `/api/page-create`;
export const postPageRefresh = `/api/page-refresh`;
export const listPages = `/api/pages`;
export const listPagesByFTS = `/api/pages-by-fts`;
export const getSharedArticle = `/api/shared-article`;
export const postSharedArticle = `/api/shared-article`;
export const postSharedArticleFromGDocs = `/api/shared-article-from-gdocs`;
export const listSharedArticles = `/api/shared-articles`;
export const getSidebar = `/api/sidebar`;
export const saveSidebar = `/api/sidebar`;
export const getSiteData = `/api/site-data`;
export const postSiteData = `/api/site-data`;
export const getSiteParams = `/api/site-params`;
export const postSiteParams = `/api/site-params`;
