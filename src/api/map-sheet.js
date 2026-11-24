import { getMapSheet, get } from "./client.ts";
import { processGDocsDoc } from "./gdocs.js";

export async function prepareMapSheet(id) {
  let [info, err] = await get(getMapSheet, { id });
  if (err) {
    return [null, err];
  }
  if (info.featured_doc_id) {
    [, err] = await processGDocsDoc(info.featured_doc_id);
    if (err) {
      return [null, err];
    }
  }
  return [info, null];
}
