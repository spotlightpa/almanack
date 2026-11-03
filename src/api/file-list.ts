import { ref, computed } from "vue";

import { get, post, listFiles, updateFile, uploadFile } from "./client.ts";
import { makeState } from "@/api/loader.ts";

interface FileEntry {
  url: string;
  description: string;
}

interface FileListResponse {
  files: FileEntry[];
}

export function useFileList() {
  let { apiStateRefs, exec } = makeState();

  const files = computed(
    () => (apiStateRefs.rawData.value as FileListResponse | null)?.files ?? []
  );
  const isDragging = ref(false);
  const isUploading = ref(false);
  const uploadError = ref<unknown>(null);
  const fileURL = ref<string | null>(null);

  let actions = {
    async fetch() {
      exec(() => get<FileListResponse>(listFiles));
    },
    updateDescription(file: FileEntry) {
      let description = window.prompt("Update description", file.description);
      if (description !== null && description !== file.description) {
        exec(() =>
          post(updateFile, {
            url: file.url,
            description,
            set_description: true,
          }).then(() => get<FileListResponse>(listFiles))
        );
      }
    },
    async uploadFileInput(ev: { target: { files: FileList | File[] } }) {
      let { files: inputFiles } = ev.target;
      isUploading.value = true;
      uploadError.value = null;

      for (let body of inputFiles) {
        [fileURL.value, uploadError.value] = await uploadFile(body);
        if (uploadError.value) {
          break;
        }
      }
      isUploading.value = false;
      await actions.fetch();
    },
    dropFile(ev: DragEvent) {
      isDragging.value = false;
      let files = ev.dataTransfer?.files ?? [];
      return actions.uploadFileInput({ target: { files } });
    },
  };

  actions.fetch();

  return {
    ...apiStateRefs,
    files,
    isDragging,
    isUploading,
    uploadError,
    fileURL,
    ...actions,
  };
}
