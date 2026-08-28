declare module "@/api/client-v2.js" {
  export function get(
    url: string,
    params?: Record<string, unknown>
  ): Promise<[unknown, unknown]>;
  export function post(url: string, obj?: unknown): Promise<[unknown, unknown]>;

  export const listAllSeries: string;
  export const listAllTopics: string;
  export const postAuthorizedDomain: string;
  export const listAuthorizedDomains: string;
  export const postAuthorizedEmailAddress: string;
  export const listAuthorizedEmailAddresses: string;
  export const createSignedUpload: string;
  export const postDonorWall: string;
  export const createFile: string;
  export const listFiles: string;
  export const updateFile: string;
  export const getGDocsDoc: string;
  export const postGDocsDoc: string;
  export const postImageUpdate: string;
  export const listImages: string;
  export const sendMessage: string;
  export const getPage: string;
  export const postPage: string;
  export const postPageJSON: string;
  export const postPageCreate: string;
  export const postPageLoad: string;
  export const postPageRefresh: string;
  export const listPages: string;
  export const listPagesByFTS: string;
  export const listPromotions: string;
  export const postPromotion: string;
  export const deletePromotion: string;
  export const getSharedArticle: string;
  export const postSharedArticle: string;
  export const postSharedArticleFromGDocs: string;
  export const listSharedArticles: string;
  export const getSidebar: string;
  export const saveSidebar: string;
  export const getSiteData: string;
  export const postSiteData: string;
  export const getSiteParams: string;
  export const postSiteParams: string;
}
