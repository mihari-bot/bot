package chardef

type Chardef struct {
	Name             string   `yaml:"name"`
	Alias            []string `yaml:"alias"`
	Age              string   `yaml:"age"`
	Gender           string   `yaml:"gender"`
	Appearance       string   `yaml:"appearance"`
	Worldview        string   `yaml:"worldview"`
	Personality      []string `yaml:"personality"`
	SpeechStyle      string   `yaml:"speech_style"`
	Catchphrase      string   `yaml:"catchphrase"`
	EmotionRange     []string `yaml:"emotion_range"`
	Backstory        string   `yaml:"backstory"`
	Relationship     string   `yaml:"relationship"`
	RoleGoal         string   `yaml:"role_goal"`
	CanDo            []string `yaml:"can_do"`
	MustDo           []string `yaml:"must_do"`
	MustNotDo        []string `yaml:"must_not_do"`
	SoftBoundaries   []string `yaml:"soft_boundaries"`
	ReplyLength      string   `yaml:"reply_length"`
	LanguageMix      []string `yaml:"language_mix"`
	UseActionDesc    bool     `yaml:"use_action_desc"`
	EmojiStyle       string   `yaml:"emoji_style"`
	MarkdownEnabled  bool     `yaml:"markdown_enabled"`
	TriggerWords     []string `yaml:"trigger_words"`
	SecretInfo       string   `yaml:"secret_info"`
	ExampleDialogues []struct {
		UserInput string `yaml:"user_input"`
		CharReply string `yaml:"char_reply"`
		Note      string `yaml:"note"`
	} `yaml:"example_dialogues"`
}
