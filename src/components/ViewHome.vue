<template>
  <div class="section container content">
    <h2>
      Hello {{ $auth.user.user_metadata.full_name }} (<span
        v-for="role of $auth.user.app_metadata.roles"
        :key="role"
        v-text="role"
      ></span
      >).
    </h2>
    <h2>Token:</h2>
    <pre><code>{{ $auth.token }}</code></pre>
  </div>
</template>
