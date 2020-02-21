<script>
import AdminListPanel from "./AdminListPanel.vue";

export default {
  name: "AdminList",
  components: {
    AdminListPanel,
  },
  props: {
    articles: Array,
    title: String,
  },
};
</script>
<template>
  <nav class="panel is-black">
    <h1 class="panel-heading" v-text="title"></h1>
    <article v-for="article of articles" :key="article.id" class="panel-block">
      <AdminListPanel :article="article" @refresh="$emit('refresh', $event)" />
    </article>
  </nav>
</template>
