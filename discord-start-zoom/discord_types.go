package function

// DiscordInteraction is the request we get from Discord when a user
// triggers a slash Command i.e. /zoom
type DiscordInteraction struct {
	Type   float64                  `json:"type"`
	Data   DiscordInteractionData   `json:"data"`
	Member DiscordInteractionMember `json:"member"`
}

// DiscordInteractionData is present for the slash command itself
// i.e. /zoom
type DiscordInteractionData struct {
	Name    string                          `json:"name"`
	ID      string                          `json:"id"`
	Type    float64                         `json:"type"`
	Options []DiscordInteractionDataOptions `json:"options"`
}

// DiscordInteractionDataOptions contains the option passed in
// within the slash command i.e. the parameters
type DiscordInteractionDataOptions struct {
	Name  string      `json:"name"`
	Type  float64     `json:"type"`
	Value interface{} `json:"value"`
}

type DiscordInteractionMember struct {
	User DiscordInteractionMemberUser `json:"user"`
}

// DiscordInteractionMemberUser gives a way to uniquely
// identify a user by adding # between the Username and
// the Discriminator
type DiscordInteractionMemberUser struct {
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
}

// DiscordResponse is the response we send back to Discord
// See also: https://discord.com/developers/docs/interactions/receiving-and-responding
type DiscordResponse struct {
	Type float64             `json:"type"`
	Data DiscordResponseData `json:"data"`
}

type DiscordResponseData struct {
	Content string                 `json:"content"`
	Embeds  []DiscordResponseEmbed `json:"embeds,omitempty"`
}

type DiscordResponseEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Type        string `json:"type"`
}
