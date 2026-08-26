<script setup>
import { ref } from "vue";

import { useAuth } from "@/api/auth.js";

const auth = useAuth();

const email = ref(auth.email.value);
const fullName = ref(auth.fullName.value);
const selectedRoles = ref([...auth.roles.value]);

const ROLES = ["admin", "Spotlight PA", "editor", "arc user"];

function applyLogin() {
  auth.setDevUser({
    email: email.value,
    fullName: fullName.value,
    roles: selectedRoles.value,
  });
}
</script>

<template>
  <MetaHead>
    <title>Dev Login • Spotlight PA Almanack</title>
  </MetaHead>

  <h1 class="title">Dev Login</h1>

  <div class="field">
    <label class="label">Email</label>
    <div class="control">
      <input v-model="email" class="input" type="email" />
    </div>
  </div>

  <div class="field">
    <label class="label">Full name</label>
    <div class="control">
      <input v-model="fullName" class="input" type="text" />
    </div>
  </div>

  <div class="field">
    <label class="label">Roles</label>
    <div class="control">
      <label v-for="role in ROLES" :key="role" class="checkbox mr-4">
        <input v-model="selectedRoles" type="checkbox" :value="role" />
        {{ role }}
      </label>
    </div>
  </div>

  <div class="field is-grouped">
    <div class="control">
      <button
        class="button is-primary has-text-weight-semibold"
        @click="applyLogin"
      >
        Log in
      </button>
    </div>
    <div class="control">
      <button
        class="button is-light has-text-weight-semibold"
        @click="auth.logout"
      >
        Log out
      </button>
    </div>
  </div>
</template>
