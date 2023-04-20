import { getGDocsDoc, postGDocsDoc, get, post } from "./client-v2";

function wait(milliseconds) {
  return new Promise((resolve) => {
    window.setTimeout(resolve, milliseconds);
  });
}

export async function processGDocsDoc(gdocsDocID) {
  // Create job
  let [dbDoc, err] = await post(postGDocsDoc, { gdocs_id: gdocsDocID });
  if (err) {
    return [null, err];
  }
  // Kick off task runner
  try {
    await window.fetch("/api-background/images");
  } catch (err) {
    return [null, err];
  }

  // Poll while waiting for task to complete
  while (!dbDoc.processed_at) {
    await wait(250);
    [dbDoc, err] = await get(getGDocsDoc, { id: dbDoc.id });
    if (err) {
      return [null, err];
    }
  }
  return [dbDoc, null];
}
