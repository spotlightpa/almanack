import { reactive, computed, toRefs } from "vue";

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

  const state = reactive({
    files: computed(() => {
      return (
        (apiStateRefs.rawData.value as FileListResponse | null)?.files ?? []
      );
    }),
    isDragging: false,
    isUploading: false,
    uploadError: null as unknown,
    fileURL: null as string | null,
  });

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
      let { files } = ev.target;
      state.isUploading = true;
      state.uploadError = null;

      for (let body of files) {
        [state.fileURL, state.uploadError] = await uploadFile(body);
        if (state.uploadError) {
          break;
        }
      }
      state.isUploading = false;
      await actions.fetch();
    },
    dropFile(ev: DragEvent) {
      state.isDragging = false;
      let files = ev.dataTransfer?.files ?? [];
      return actions.uploadFileInput({ target: { files } });
    },
  };

  actions.fetch();

  return {
    ...apiStateRefs,
    ...toRefs(state),
    ...actions,
  };
}
