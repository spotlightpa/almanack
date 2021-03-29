<script>
import { computed, reactive, ref, toRefs } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";

import APILoader from "./APILoader.vue";

export default {
  name: "ViewAuthorizedDomains",
  components: {
    APILoader,
  },
  metaInfo: {
    title: "Authorized Domains and Addresses",
  },
  setup() {
    let { listAuthorizedDomains, addAuthorizedDomain } = useClient();

    let { apiState: listState, exec: listExec } = makeState();
    let { apiState: addState, exec: addExec } = makeState();

    let apiState = reactive({
      didLoad: computed(() => listState.didLoad || addState.didLoad),
      isLoading: computed(() => listState.isLoading || addState.isLoading),
      error: computed(() => listState.error || addState.error),
      domains: computed(() => listState.rawData?.domains || []),
    });

    async function list() {
      await listExec(listAuthorizedDomains);
    }

    async function addDomain(domain) {
      await addExec(() => addAuthorizedDomain({ domain, remove: false }));
      await list();
    }

    async function removeDomain(domain) {
      if (!window.confirm(`Are you sure you want to remove ${domain}?`)) {
        return;
      }
      await addExec(() => addAuthorizedDomain({ domain, remove: true }));
      await list();
    }

    list();

    return {
      name: ref(""),
      list,
      addDomain,
      removeDomain,
      ...toRefs(apiState),
    };
  },
};
</script>

<template>
  <div>
    <h2 class="title">Authorized domains</h2>
    <APILoader :is-loading="isLoading" :reload="list" :error="error">
      <div class="field is-grouped is-grouped-multiline">
        <div v-for="domain of domains" :key="domain" class="control">
          <div class="tags has-addons">
            <span class="tag is-small" v-text="domain" />
            <a class="tag is-delete" @click="removeDomain(domain)" />
          </div>
        </div>
      </div>
      <form class="field has-addons" @submit.prevent="addDomain(name)">
        <div class="control is-expanded">
          <input v-model="name" class="input" />
        </div>
        <div class="control">
          <button
            class="button has-text-weight-semibold is-primary"
            :class="{ 'is-loading': isLoading }"
            @click="
              addDomain(name).then(() => {
                name = '';
              })
            "
          >
            Add domain
          </button>
        </div>
      </form>
    </APILoader>
  </div>
</template>
