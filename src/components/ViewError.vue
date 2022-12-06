<script>
export default {
  props: { error: { default: null } },
};
</script>

<template>
  <MetaHead>
    <title>Error • Spotlight PA</title>
  </MetaHead>

  <div class="content is-large">
    <h2>Error</h2>
    <p><a href="/">Go home</a>?</p>

    <div v-if="error" class="message is-danger">
      <p class="message-header">{{ error.name }}</p>
      <p class="message-body">{{ error.message }}</p>
    </div>
  </div>
</template>
