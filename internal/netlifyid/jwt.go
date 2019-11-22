package netlifyid

type JWT struct {
    Identity Identity `json:"identity"`
    SiteURL  string   `json:"site_url"`
    User     User     `json:"user"`
}
type Identity struct {
    Token string `json:"token"`
    URL   string `json:"url"`
}
type AppMetadata struct {
    Provider string   `json:"provider"`
    Roles    []string `json:"roles"`
}
type UserMetadata struct {
    FullName string `json:"full_name"`
}
type User struct {
    AppMetadata  AppMetadata  `json:"app_metadata"`
    Email        string       `json:"email"`
    Exp          int          `json:"exp"`
    Sub          string       `json:"sub"`
    UserMetadata UserMetadata `json:"user_metadata"`
}
