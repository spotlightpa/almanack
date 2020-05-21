<script>
import { computed, reactive, ref, toRefs } from "@vue/composition-api";

import { useAuth, useClient, makeState } from "@/api/hooks.js";

import APILoader from "./APILoader.vue";

export default {
  name: "ViewAuthorizedDomains",
  components: {
    APILoader,
  },
  metaInfo: {
    title: "Authorized Domains",
  },
  setup() {
    let { isSpotlightPAUser } = useAuth();
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
      await addExec(() => addAuthorizedDomain({ domain }));
      await list();
    }

    list();

    return {
      isSpotlightPAUser,

      name: ref(""),
      list,
      addDomain,
      ...toRefs(apiState),
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
      <APILoader
        :can-load="isSpotlightPAUser"
        :is-loading="isLoading"
        :reload="list"
        :error="error"
      >
        <h2 class="title">Authorized domains</h2>
        <ul class="tags">
          <li
            v-for="domain of domains"
            :key="domain"
            class="tag"
            v-text="domain"
          />
        </ul>
        <form class="field has-addons" @submit.prevent="addDomain(name)">
          <div class="control is-expanded">
            <input v-model="name" class="input" />
          </div>
          <div class="control">
            <button
              class="button has-text-weight-semibold is-primary"
              :class="{ 'is-loading': isLoading }"
              @click="addDomain(name)"
            >
              Add domain
            </button>
          </div>
        </form>
      </APILoader>
    </div>
  </div>
</template>
