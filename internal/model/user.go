package model

type User struct {
	Email    string
	Password string
}

type Role string

func (r Role) String() string {
	return string(r)
}

var (
	RoleUser             Role = "user"
	RoleAssistant        Role = "assistant"
	RoleTechnical        Role = "technical"
	RoleDocumentRedirect Role = "document_redirect"
	RoleExtraQuestions   Role = "extra_questions"
)

type MessageStatus string

var (
	StatusFail MessageStatus = "FAIL"
	StatusOk   MessageStatus = "OK"
)

func (r MessageStatus) String() string {
	return string(r)
}

func GetRole(fromChatBot bool) Role {
	if fromChatBot {
		return RoleAssistant
	}
	return RoleUser
}
