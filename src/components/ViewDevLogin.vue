<script>
import { ref } from "vue";
import { useAuth } from "@/api/auth.js";

export default {
  setup() {
    const auth = useAuth();

    const email = ref(auth.email.value || "");
    const fullName = ref(auth.fullName.value || "");

    const AVAILABLE_ROLES = ["admin", "Spotlight PA", "editor", "arc user"];
    const selectedRoles = ref([...auth.roles.value]);

    function applyLogin() {
      auth.setDevUser({
        email: email.value,
        fullName: fullName.value,
        roles: selectedRoles.value,
      });
    }

    function applyLogout() {
      auth.logout();
      email.value = "";
      fullName.value = "";
      selectedRoles.value = [];
    }

    return {
      email,
      fullName,
      selectedRoles,
      AVAILABLE_ROLES,
      isSignedIn: auth.isSignedIn,
      currentEmail: auth.email,
      currentRoles: auth.roles,
      applyLogin,
      applyLogout,
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Dev Login • Almanack</title>
  </MetaHead>
  <div class="section">
    <div class="container" style="max-width: 480px">
      <div class="box">
        <h1 class="title is-4">🛠 Dev Login</h1>
        <p class="subtitle is-6 has-text-grey">
          Development-only. Sets fake auth state in
          <code>localStorage</code> without a real Netlify Identity session.
        </p>

        <div v-if="isSignedIn" class="notification is-success is-light mb-4">
          <strong>Signed in as:</strong> {{ currentEmail }}<br />
          <strong>Roles:</strong>
          {{ currentRoles.length ? currentRoles.join(", ") : "(none)" }}
        </div>
        <div v-else class="notification is-warning is-light mb-4">
          Not signed in.
        </div>

        <div class="field">
          <label class="label">Email</label>
          <div class="control">
            <input
              v-model="email"
              class="input"
              type="email"
              placeholder="you@example.com"
            />
          </div>
        </div>

        <div class="field">
          <label class="label">Full name</label>
          <div class="control">
            <input
              v-model="fullName"
              class="input"
              type="text"
              placeholder="Jane Doe"
            />
          </div>
        </div>

        <div class="field">
          <label class="label">Roles</label>
          <div class="control">
            <label
              v-for="role in AVAILABLE_ROLES"
              :key="role"
              class="checkbox mr-4"
            >
              <input v-model="selectedRoles" type="checkbox" :value="role" />
              {{ role }}
            </label>
          </div>
        </div>

        <div class="field is-grouped mt-5">
          <div class="control">
            <button class="button is-primary" @click="applyLogin">
              Apply login
            </button>
          </div>
          <div class="control">
            <button class="button is-light" @click="applyLogout">
              Log out
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
