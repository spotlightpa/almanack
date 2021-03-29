<script>
import { computed, reactive, ref, toRefs } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";

import APILoader from "./APILoader.vue";

function domainState() {
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
    list,
    addDomain,
    removeDomain,
    ...toRefs(apiState),
  };
}

export default {
  name: "ViewAuthorizedDomains",
  components: {
    APILoader,
  },
  metaInfo: {
    title: "Authorized Domains and Addresses",
  },
  setup() {
    return {
      newDomainName: ref(""),
      domain: domainState(),
    };
  },
};
</script>

<template>
  <div>
    <h2 class="title">Authorized domains</h2>
    <APILoader
      :is-loading="domain.isLoading.value"
      :reload="domain.list.value"
      :error="domain.error.value"
    >
      <div class="field is-grouped is-grouped-multiline">
        <div v-for="d of domain.domains.value" :key="d" class="control">
          <div class="tags has-addons">
            <span class="tag is-small" v-text="d" />
            <a class="tag is-delete" @click="domain.removeDomain(d)" />
          </div>
        </div>
      </div>
    </APILoader>
    <form
      class="mt-5 field has-addons"
      @submit.prevent="
        domain.addDomain(newDomainName).then(() => {
          newDomainName = '';
        })
      "
    >
      <div class="control is-expanded">
        <input v-model="newDomainName" class="input" />
      </div>
      <div class="control">
        <button
          class="button has-text-weight-semibold is-primary"
          :class="{ 'is-loading': domain.isLoading.value }"
        >
          Add domain
        </button>
      </div>
    </form>
  </div>
</template>
