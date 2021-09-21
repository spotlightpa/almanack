<script>
import { formatDate, formatTime } from "@/utils/time-format.js";

import BulmaField from "./BulmaField.vue";

export default {
  components: {
    BulmaField,
  },
  props: { files: Array },
  setup() {
    return {
      formatDate,
      formatTime,
    };
  },
};
</script>

<template>
  <BulmaField v-if="files.length" label="Choose from recent files">
    <div class="textarea preview-frame">
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <tr v-for="file in files" :key="file.id">
            <a
              class="is-flex-tablet p-1 has-text-black"
              @click="$emit('select-file', file)"
            >
              <div>
                <div class="clamped-3">
                  <span class="has-text-weight-semibold">
                    {{ file.filename
                    }}<template v-if="file.description">: </template>
                  </span>
                  <span class="has-text-grey">
                    {{ file.description }}
                  </span>
                  <span>
                    {{ formatDate(file.created_at) }}
                  </span>
                  <span class="has-text-grey">
                    {{ formatTime(file.created_at) }}
                  </span>
                </div>
              </div>
            </a>
          </tr>
        </tbody>
      </table>
    </div>
    <p>
      <router-link :to="{ name: 'file-uploader' }" target="_blank">
        Manage files
      </router-link>
    </p>
  </BulmaField>
</template>

<style scoped>
.clamped-3 {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  overflow: hidden;
}
.preview-frame {
  height: 150px;
  overflow-y: auto;
}
</style>
