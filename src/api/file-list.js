import { reactive, computed, toRefs } from "vue";

import { get, post, listFiles, updateFile, uploadFile } from "./client-v2.js";
import { makeState } from "./service-util.js";

export function useFileList() {
  let { apiState, exec } = makeState();

  let state = reactive({
    files: computed(() => {
      return apiState.rawData?.files || [];
    }),
    isDragging: false,
    isUploading: false,
    uploadError: null,
  });

  let actions = {
    async fetch() {
      exec(() => get(listFiles));
    },
    updateDescription(file) {
      let description = window.prompt("Update description", file.description);
      if (description !== null && description !== file.description) {
        exec(() =>
          post(updateFile, {
            url: file.url,
            description,
            set_description: true,
          }).then(() => get(listFiles))
        );
      }
    },
    async uploadFileInput(ev) {
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
    dropFile(ev) {
      state.isDragging = false;
      let { files = [] } = ev.dataTransfer;
      return actions.uploadFileInput({ target: { files } });
    },
  };

  actions.fetch();

  return {
    ...toRefs(apiState),
    ...toRefs(state),
    ...actions,
  };
}
