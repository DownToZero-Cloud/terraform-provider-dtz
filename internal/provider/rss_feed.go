package provider

type RssFeed struct {
	Id            string `tfsdk:"id"`
	Url           string `tfsdk:"url"`
	LastCheck     string `tfsdk:"lastCheck"`
	LastDataFound string `tfsdk:"lastDataFound"`
	Enabled       bool   `tfsdk:"enabled"`
	Name          string `tfsdk:"name"`
}
