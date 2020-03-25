<script>
import { useAuth } from "@/api/hooks.js";

import ImageUploader from "./ImageUploader.vue";

export default {
  name: "ViewUploader",
  components: {
    ImageUploader,
  },
  metaInfo: {
    title: "Upload Image",
  },
  setup() {
    let { isSpotlightPAUser } = useAuth();

    return {
      isSpotlightPAUser,
    };
  },
};
</script>

<template>
  <div>
    <div v-if="!isSpotlightPAUser" class="message is-danger">
      <p class="message-header">Not Authorized</p>

      <p class="message-body">
        You do not have permission to use this page.
        <strong
          ><router-link :to="{ name: 'home' }">Go home</router-link>?</strong
        >
      </p>
    </div>
    <div v-else>
      <h1 class="title">Upload an image</h1>
      <ImageUploader />
    </div>
  </div>
</template>
