import Vue from "vue";
import Router from "vue-router";
import { watch } from "@vue/composition-api";

import { useAuth } from "@/api/hooks.js";
import { setDimensions, sendGAPageview } from "@/utils/google-analytics.js";

import ViewAdmin from "./components/ViewAdmin.vue";
import ViewArcArticle from "./components/ViewArcArticle.vue";
import ViewAuthorizedDomains from "./components/ViewAuthorizedDomains.vue";
import ViewError from "./components/ViewError.vue";
import ViewFileUploader from "./components/ViewFileUploader.vue";
import ViewHomepageEditor from "./components/ViewHomepageEditor.vue";
import ViewLogin from "./components/ViewLogin.vue";
import ViewNewsletterList from "./components/ViewNewsletterList.vue";
import ViewNewsletterPage from "./components/ViewNewsletterPage.vue";
import ViewNewsPage from "./components/ViewNewsPage.vue";
import ViewSharedList from "./components/ViewSharedList.vue";
import ViewNewsPagesList from "./components/ViewNewsPagesList.vue";
import ViewRedirectArcToNewsPage from "./components/ViewRedirectArcToNewsPage.vue";
import ViewSidebarItems from "./components/ViewSidebarItems.vue";
import ViewSiteParams from "./components/ViewSiteParams.vue";
import ViewUnauthorized from "./components/ViewUnauthorized.vue";
import ViewUploader from "./components/ViewUploader.vue";

Vue.use(Router);

let { roles, fullName, email, isEditor, isSpotlightPAUser, isSignedIn } =
  useAuth();

let router = new Router({
  mode: "history",
  base: process.env.BASE_URL,
  routes: [
    {
      path: "/",
      name: "home",
      redirect: () => {
        if (isSpotlightPAUser.value) {
          return { name: "admin" };
        }
        if (isEditor.value) {
          return { name: "articles" };
        }
        if (isSignedIn.value) {
          return { name: "unauthorized" };
        }
        return { name: "login" };
      },
    },
    {
      path: "/login",
      name: "login",
      component: ViewLogin,
      meta: {},
    },
    {
      path: "/unauthorized",
      name: "unauthorized",
      component: ViewUnauthorized,
      meta: {
        requiresAuth: isSignedIn,
      },
    },
    {
      path: "/articles",
      name: "articles",
      component: ViewSharedList,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isEditor,
      },
    },
    {
      path: "/articles/:id",
      name: "article",
      component: ViewArcArticle,
      props: true,
      meta: { requiresAuth: isEditor },
    },
    {
      path: "/admin",
      name: "admin",
      component: ViewAdmin,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/site-params",
      name: "site-params",
      component: ViewSiteParams,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/editors-picks",
      name: "editors-picks",
      component: ViewHomepageEditor,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/sidebar-items",
      name: "sidebar-items",
      component: ViewSidebarItems,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/uploader",
      name: "uploader",
      component: ViewUploader,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/file-uploader",
      name: "file-uploader",
      component: ViewFileUploader,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/news",
      name: "news-pages",
      component: ViewNewsPagesList,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/redirect/arc-to-news/:id",
      name: "redirect-arc-news-page",
      component: ViewRedirectArcToNewsPage,
      props: true,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/news/:id",
      name: "news-page",
      component: ViewNewsPage,
      props: true,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/domains",
      name: "domains",
      component: ViewAuthorizedDomains,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/newsletters",
      name: "newsletters",
      component: ViewNewsletterList,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/newsletters/:id",
      name: "newsletter-page",
      component: ViewNewsletterPage,
      props: true,
      meta: { requiresAuth: isSpotlightPAUser },
    },
    {
      path: "/*",
      name: "error",
      component: ViewError,
    },
  ],
  scrollBehavior(to, from, savedPosition) {
    return savedPosition || { selector: "#top-nav" };
  },
});

router.beforeEach((to, from, next) => {
  let record = to.matched.find((record) => record.meta.requiresAuth);
  if (record) {
    if (!record.meta.requiresAuth.value) {
      let redirect = to.fullPath || from.fullPath;
      next({
        name: "login",
        hash: to.hash, // For verifying tokens etc.
        replace: true,
        query: { redirect },
      });
      return;
    }
  }
  next();
});

router.afterEach((to) => {
  let domain = /@(.*)$/.exec(email.value)?.[1].toLowerCase() ?? "None";
  let name = fullName.value || "Not Signed In";
  let role =
    roles.value.find((r) => r === "admin") ||
    roles.value.find((r) => r === "Spotlight PA") ||
    roles.value.find((r) => r === "editor") ||
    "None";

  setDimensions({
    domain,
    name,
    role,
  });
  sendGAPageview(to.fullPath);
});

watch(
  () => isSignedIn.value,
  (newStatus, oldStatus) => {
    if (oldStatus === undefined || newStatus === oldStatus) {
      return;
    }
    let destination = { name: "home" };
    if (newStatus && router.app.$route.query?.redirect) {
      let path = router.app.$route.query.redirect;
      let { route } = router.resolve({ path });
      if (route) {
        if (!route.meta.requiresAuth || route.meta.requiresAuth.value) {
          destination = { path };
        }
      }
    }
    // Use a timeout because isSpotlight, etc won't be updated yet when push runs
    window.setTimeout(() => router.push(destination), 0);
  },
  {
    immediate: true,
  }
);

export default router;
