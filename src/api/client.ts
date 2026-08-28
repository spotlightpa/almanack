// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore – auth.js has no types yet
import { useAuth } from "./auth.js";

/** [data, null] on success; [falsy, error] on failure */
type Result<T> = [T, null] | [unknown, Error];

const tryTo = <T>(promise: Promise<T>): Promise<Result<T>> =>
  promise
    // Wrap data/errors
    .then((data): [T, null] => [data, null])
    .catch((error: Error): [unknown, Error] => [null, error]);

interface ResponseErrorDetails {
  [key: string]: unknown;
}

interface AppError extends Error {
  details: ResponseErrorDetails;
}

const responseError = async (
  rsp: Response
): Promise<AppError | undefined> => {
  if (rsp.ok) {
    return;
  }
  let details: ResponseErrorDetails = {};
  try {
    details = ((await rsp.json()) as { details?: ResponseErrorDetails })
      ?.details ?? {};
    // eslint-disable-next-line no-empty
  } catch (e) {}

  let msg = `${rsp.status} ${rsp.statusText}`;
  let err = new Error("Unexpected response from server: " + msg) as AppError;
  err.name = msg;
  err.details = details;
  return err;
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
let $auth: any = null;

interface RequestOptions {
  params?: Record<string, string> | null;
  headers?: Record<string, string>;
  options?: RequestInit;
}

async function request<T = unknown>(
  url: string,
  { params = null, headers = {}, options = {} }: RequestOptions = {}
): Promise<T> {
  if (!$auth) {
    $auth = useAuth();
  }

  let authHeaders = await $auth.headers();
  if (!authHeaders) {
    let err = new Error("Please log in again.") as AppError;
    err.name = "Login Error";
    err.details = {};
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

  return (await resp.json()) as T;
}

export function get<T = unknown>(
  url: string,
  params?: Record<string, string>
): Promise<Result<T>> {
  return tryTo(request<T>(url, { params }));
}

export function post<T = unknown>(
  url: string,
  obj: unknown
): Promise<Result<T>> {
  let body = JSON.stringify(obj);
  return tryTo(
    request<T>(url, {
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
export const postPageJSON = `/api/page-json`;
export const postPageCreate = `/api/page-create`;
export const postPageLoad = `/api/page-load`;
export const postPageRefresh = `/api/page-refresh`;
export const listPages = `/api/pages`;
export const listPagesByFTS = `/api/pages-by-fts`;
export const listPromotions = `/api/promotion`;
export const postPromotion = `/api/promotion`;
export const deletePromotion = `/api/promotion-delete`;
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

interface SignedUploadResponse {
  "signed-url": string;
  filename: string;
}

export async function uploadImage(body: File): Promise<Result<string>> {
  let [data, err] = await post<SignedUploadResponse>(createSignedUpload, {
    type: body.type,
  });
  if (err) {
    return ["", err];
  }
  let { "signed-url": signedURL, filename } = data as SignedUploadResponse;
  let rsp: Response;
  try {
    rsp = await fetch(signedURL, { method: "PUT", body });
  } catch (e) {
    return ["", e as Error];
  }
  if (!rsp.ok) {
    return ["", (await responseError(rsp))!];
  }
  [, err] = await post(postImageUpdate, {
    path: filename,
    set_credit: false,
    set_description: false,
  });
  if (err) {
    return ["", err];
  }
  return [filename, null];
}

interface CreateFileResponse {
  "signed-url": string;
  "file-url": string;
  "cache-control": string;
  disposition: string;
}

export async function uploadFile(body: File): Promise<Result<string>> {
  let [data, err] = await post<CreateFileResponse>(createFile, {
    filename: body.name,
    mimeType: body.type,
  });
  if (err) {
    return ["", err];
  }
  let {
    "signed-url": signedURL,
    "file-url": fileURL,
    "cache-control": cacheControl,
    disposition,
  } = data as CreateFileResponse;
  let opts: RequestInit = {
    method: "PUT",
    body,
    headers: {
      "Content-Disposition": disposition,
      "Cache-Control": cacheControl,
    },
  };
  let rsp: Response;
  try {
    rsp = await fetch(signedURL, opts);
  } catch (e) {
    return ["", e as Error];
  }
  if (!rsp.ok) {
    return ["", (await responseError(rsp))!];
  }
  [, err] = await post(updateFile, {
    url: fileURL,
    description: null,
    set_description: false,
  });
  if (err) {
    return ["", err];
  }
  return [fileURL, null];
}
