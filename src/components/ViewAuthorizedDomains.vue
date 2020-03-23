<script>
import { reactive, ref, toRefs } from "@vue/composition-api";

import { useAuth, useClient } from "@/api/hooks.js";

import APILoader from "./APILoader.vue";

export default {
  name: "ViewAuthorizedDomains",
  components: {
    APILoader,
  },
  setup() {
    let { isSpotlightPAUser } = useAuth();
    let { listAuthorizedDomains, addAuthorizedDomain } = useClient();
    let apiState = reactive({
      didLoad: false,
      isLoading: false,
      domains: [],
      error: null,
    });

    async function list() {
      apiState.isLoading = true;
      let data;
      [data, apiState.error] = await listAuthorizedDomains();
      apiState.domains = data.domains;
      apiState.didLoad = true;
      apiState.isLoading = false;
    }

    async function addDomain(domain) {
      console.log(domain);
      apiState.isLoading = true;
      [, apiState.error] = await addAuthorizedDomain(domain);
      if (apiState.error) {
        apiState.isLoading = false;
        return;
      }
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
        <ul class="list is-hoverable">
          <li
            v-for="domain of domains"
            :key="domain"
            class="list-item"
            v-text="domain"
          ></li>
        </ul>
        <div class="field has-addons">
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
        </div>
      </APILoader>
    </div>
  </div>
</template>
