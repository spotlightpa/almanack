<script>
export default {
  props: {
    title: {
      type: String,
      required: true,
    },
    page: {
      type: String,
    },
    nextPage: {
      type: Object,
    },
    apiState: {
      type: Object,
      required: true,
    },
    reload: {
      type: Function,
      required: true,
    },
    pages: {
      type: Array,
      required: true,
    },
  },
  setup(props) {
    return {
      ...props.apiState,
    };
  },
};
</script>

<template>
  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: title, to: {} },
      ]"
    />
    <h1 class="title">
      {{ title }}
      <template v-if="page">(overflow page {{ page }})</template>
    </h1>
    <PageLookup />
    <APILoader :is-loading="isLoading" :reload="reload" :error="error">
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <PageListRow
            v-for="page of pages"
            :key="page.id"
            :link="page.link"
            :status="page.status"
            :label="page.internalID"
            :date="page.publicationDate"
            :hed="page.title"
            :dek="page.blurb"
            :image="page.image"
            :image-alt="page.image"
          />
        </tbody>
      </table>

      <div class="buttons mt-5">
        <router-link
          v-if="nextPage"
          :to="nextPage"
          class="button is-primary has-text-weight-semibold"
        >
          Show Older Stories…
        </router-link>
      </div>
    </APILoader>
  </div>
</template>
