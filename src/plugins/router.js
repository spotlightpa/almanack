import { defineAsyncComponent, nextTick, watch } from "vue";
import { createRouter, createWebHistory } from "vue-router";

import { useAuth } from "@/api/hooks.js";
import { setDimensions, sendGAPageview } from "@/utils/google-analytics.js";

import ViewError from "@/components/ViewError.vue";

let viewNames = [
  "ViewAdmin",
  "ViewArcArticle",
  "ViewAuthorizedDomains",
  "ViewFileUploader",
  "ViewHomepageEditor",
  "ViewImageUploader",
  "ViewLogin",
  "ViewNewsletterList",
  "ViewNewsletterPage",
  "ViewNewsPage",
  "ViewNewsPagesList",
  "ViewRedirectArcToNewsPage",
  "ViewSharedList",
  "ViewSidebarItems",
  "ViewSiteParams",
  "ViewStateCollegeEditor",
  "ViewUnauthorized",
];
let viewComponents = {};
for (let name of viewNames) {
  viewComponents[name] = defineAsyncComponent(() =>
    import(`@/components/${name}.vue`)
  );
}

let { roles, fullName, email, isEditor, isSpotlightPAUser, isSignedIn } =
  useAuth();

let router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
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
      component: viewComponents.ViewLogin,
      meta: {},
    },
    {
      path: "/unauthorized",
      name: "unauthorized",
      component: viewComponents.ViewUnauthorized,
      meta: {
        requiresAuth: isSignedIn,
      },
    },
    {
      path: "/articles",
      name: "articles",
      component: viewComponents.ViewSharedList,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isEditor,
      },
    },
    {
      path: "/articles/:id",
      name: "article",
      component: viewComponents.ViewArcArticle,
      props: true,
      meta: { requiresAuth: isEditor },
    },
    {
      path: "/admin",
      name: "admin",
      component: viewComponents.ViewAdmin,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/site-params",
      name: "site-params",
      component: viewComponents.ViewSiteParams,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/editors-picks",
      name: "homepage-editor",
      component: viewComponents.ViewHomepageEditor,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/state-college-editor",
      name: "state-college-editor",
      component: viewComponents.ViewStateCollegeEditor,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/sidebar-items",
      name: "sidebar-items",
      component: viewComponents.ViewSidebarItems,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/uploader",
      name: "image-uploader",
      component: viewComponents.ViewImageUploader,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/file-uploader",
      name: "file-uploader",
      component: viewComponents.ViewFileUploader,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/news",
      name: "news-pages",
      component: viewComponents.ViewNewsPagesList,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/statecollege",
      name: "statecollege-pages",
      component: viewComponents.ViewStateCollegeList,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/redirect/arc-to-news/:id",
      name: "redirect-arc-news-page",
      component: viewComponents.ViewRedirectArcToNewsPage,
      props: true,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/news/:id",
      name: "news-page",
      component: viewComponents.ViewNewsPage,
      props: true,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/domains",
      name: "domains",
      component: viewComponents.ViewAuthorizedDomains,
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/newsletters",
      name: "newsletters",
      component: viewComponents.ViewNewsletterList,
      props: (route) => ({ page: route.query.page }),
      meta: {
        requiresAuth: isSpotlightPAUser,
      },
    },
    {
      path: "/admin/newsletters/:id",
      name: "newsletter-page",
      component: viewComponents.ViewNewsletterPage,
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
  async (newStatus, oldStatus) => {
    if (oldStatus === undefined || newStatus === oldStatus) {
      return;
    }
    let destination = { name: "home" };
    if (newStatus && router.currentRoute.value.query?.redirect) {
      let path = router.currentRoute.value.query.redirect;
      let { matched } = router.resolve({ path });
      let [route = null] = matched;
      if (route) {
        if (!route.meta.requiresAuth || route.meta.requiresAuth.value) {
          destination = { path };
        }
      }
    }
    await nextTick();
    router.push(destination);
  },
  {
    immediate: true,
  }
);

export default router;
