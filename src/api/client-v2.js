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
export const listAllPages = `/api/all-pages`;
export const listAllSeries = `/api/all-series`;
export const listAllTopics = `/api/all-topics`;
export const postAuthorizedDomain = `/api/authorized-domains`;
export const listAuthorizedDomains = `/api/authorized-domains`;
export const postAuthorizedEmailAddress = `/api/authorized-addresses`;
export const listAuthorizedEmailAddresses = `/api/authorized-addresses`;
export const createSignedUpload = `/api/create-signed-upload`;
export const getEditorsPicks = `/api/editors-picks`;
export const saveEditorsPicks = `/api/editors-picks`;
export const getElectionFeature = `/api/election-feature`;
export const saveElectionFeature = `/api/election-feature`;
export const createFile = `/api/files-create`;
export const listFiles = `/api/files-list`;
export const updateFile = `/api/files-update`;
export const updateImage = `/api/image-update`;
export const listImages = `/api/images`;
export const getSignupURL = `/api/mailchimp-signup-url`;
export const sendMessage = `/api/message`;
export const postPage = `/api/page`;
export const getPageByFilePath = `/api/page-by-file-path`;
export const getPageByURLPath = `/api/page-by-url-path`;
export const postPageRefresh = `/api/page-refresh`;
export const listPages = `/api/pages`;
export const listPagesByFTS = `/api/pages-by-fts`;
export const getSharedArticle = `/api/shared-article`;
export const getSharedArticleBySource = `/api/shared-article-by-source`;
export const listSharedArticles = `/api/shared-articles`;
export const getSidebar = `/api/sidebar`;
export const saveSidebar = `/api/sidebar`;
export const getSiteParams = `/api/site-params`;
export const postSiteParams = `/api/site-params`;
export const getStateCollegeEditor = `/api/state-college-editor`;
export const saveStateCollegeEditor = `/api/state-college-editor`;
