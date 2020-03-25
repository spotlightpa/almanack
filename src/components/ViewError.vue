<script>
export default {
  name: "ViewError",
  metaInfo: {
    title: "Error",
  },
};
</script>

<template>
  <div class="content is-large">
    <h2>Error</h2>
    <p><router-link :to="{ name: 'home' }">Go home</router-link>?</p>
  </div>
</template>
