import { reactive, computed, toRefs } from "vue";

import { useClient } from "./client.js";
import { makeState } from "./service-util.js";

export function useFileList() {
  let { listFiles, updateFile, uploadFile } = useClient();
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
      exec(listFiles);
    },
    updateDescription(file) {
      let description = window.prompt("Update description", file.description);
      if (description !== null && description !== file.description) {
        exec(() =>
          Promise.resolve()
            .then(() => updateFile(file.url, { description }))
            .then(listFiles)
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
