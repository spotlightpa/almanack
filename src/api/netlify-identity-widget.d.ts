/** Minimal ambient types for netlify-identity-widget based on observed usage. */
declare module "netlify-identity-widget" {
  interface NetlifyUser {
    email: string;
    app_metadata: { roles?: string[] };
    user_metadata: { full_name?: string };
    token: { access_token: string } | null;
    jwt(): Promise<string>;
  }

  interface InitOptions {
    logo?: boolean;
    APIUrl?: string | null;
  }

  function init(options?: InitOptions): void;
  function open(modal: "signup" | "login"): void;
  function close(): void;
  function logout(): Promise<void>;
  function on(event: "init", cb: (user: NetlifyUser | null) => void): void;
  function on(event: "login", cb: (user: NetlifyUser) => void): void;
  function on(event: "logout", cb: () => void): void;
  function on(event: "error", cb: (err: Error) => void): void;

  const store: { user: NetlifyUser | null };
}
