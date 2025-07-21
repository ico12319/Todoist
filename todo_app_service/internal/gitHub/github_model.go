package gitHub

type Organization struct {
	Login string `json:"login"`
}
type UserInfo struct {
	Email *string `json:"email"`
}

type GitHubResponse struct {
	ParticipatesIn []Organization
	UserInfo
}
