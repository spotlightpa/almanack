<script>
import { computed, reactive, ref, toRefs } from "vue";

import { useClient, makeState } from "@/api/hooks.js";

function domainState() {
  let { listAuthorizedDomains, postAuthorizedDomain } = useClient();

  let { apiState: listState, exec: listExec } = makeState();
  let { apiState: addState, exec: addExec } = makeState();

  let apiState = reactive({
    didLoad: computed(() => listState.didLoad || addState.didLoad),
    isLoading: computed(() => listState.isLoading || addState.isLoading),
    errors: computed(() =>
      [listState.error, addState.error].filter((o) => !!o)
    ),
    domains: computed(() => listState.rawData?.domains || []),
  });

  async function list() {
    await listExec(listAuthorizedDomains);
  }

  async function addDomain(domain) {
    await addExec(() => postAuthorizedDomain({ domain, remove: false }));
    await list();
  }

  async function removeDomain(domain) {
    if (!window.confirm(`Are you sure you want to remove ${domain}?`)) {
      return;
    }
    await addExec(() => postAuthorizedDomain({ domain, remove: true }));
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

function addressState() {
  let { listAuthorizedEmailAddresses, postAuthorizedEmailAddress } =
    useClient();

  let { apiState: listState, exec: listExec } = makeState();
  let { apiState: addState, exec: addExec } = makeState();

  let apiState = reactive({
    didLoad: computed(() => listState.didLoad || addState.didLoad),
    isLoading: computed(() => listState.isLoading || addState.isLoading),
    errors: computed(() =>
      [listState.error, addState.error].filter((o) => !!o)
    ),
    addresses: computed(() => listState.rawData?.addresses || []),
  });

  async function list() {
    await listExec(listAuthorizedEmailAddresses);
  }

  async function addAddress(address) {
    await addExec(() => postAuthorizedEmailAddress({ address, remove: false }));
    await list();
  }

  async function removeAddress(address) {
    if (!window.confirm(`Are you sure you want to remove ${address}?`)) {
      return;
    }
    await addExec(() => postAuthorizedEmailAddress({ address, remove: true }));
    await list();
  }

  list();

  return {
    list,
    addAddress,
    removeAddress,
    ...toRefs(apiState),
  };
}

export default {
  setup() {
    return {
      newDomainName: ref(""),
      domain: domainState(),
      newEmailAddress: ref(""),
      address: addressState(),
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Preauthorized Domains and Addresses • Spotlight PA</title>
  </MetaHead>
  <div>
    <div>
      <h2 class="title">Preauthorized domains</h2>
      <SpinnerProgress :is-loading="domain.isLoading.value" />

      <div class="field is-grouped is-grouped-multiline">
        <div v-for="d of domain.domains.value" :key="d" class="control">
          <div class="tags has-addons">
            <span class="tag is-small" v-text="d" />
            <a class="tag is-delete" @click="domain.removeDomain(d)" />
          </div>
        </div>
      </div>

      <div v-for="error of domain.errors.value" :key="error.name">
        <ErrorReloader :error="error" @reload="domain.list()" />
      </div>

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
    <div class="mt-6">
      <h2 class="title">Preauthorized email adddresses</h2>
      <SpinnerProgress :is-loading="address.isLoading.value" />

      <div class="field is-grouped is-grouped-multiline">
        <div
          v-for="email of address.addresses.value"
          :key="email"
          class="control"
        >
          <div class="tags has-addons">
            <span class="tag is-small" v-text="email" />
            <a class="tag is-delete" @click="address.removeAddress(email)" />
          </div>
        </div>
      </div>

      <div v-for="error of address.errors.value" :key="error.name">
        <ErrorReloader :error="error" @reload="address.list()" />
      </div>

      <form
        class="mt-5 field has-addons"
        @submit.prevent="
          address.addAddress(newEmailAddress).then(() => {
            newEmailAddress = '';
          })
        "
      >
        <div class="control is-expanded">
          <input v-model="newEmailAddress" type="email" class="input" />
        </div>
        <div class="control">
          <button
            class="button has-text-weight-semibold is-primary"
            :class="{ 'is-loading': address.isLoading.value }"
          >
            Add email address
          </button>
        </div>
      </form>
    </div>
    <p class="mt-6 is-italic">
      <em class="is-bold">Note</em>: Deleting a domain or address will not
      remove any users who have already been authorized.
    </p>
  </div>
</template>
