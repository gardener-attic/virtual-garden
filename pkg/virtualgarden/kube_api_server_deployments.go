package virtualgarden

func (o *operation) getAPIServerAuditWebhookConfig() string {
	return o.imports.VirtualGarden.KubeAPIServer.AuditWebhookConfig.Config
}

func (o *operation) isSeedAuthorizerEnabled() bool {
	return o.imports.VirtualGarden.KubeAPIServer != nil && o.imports.VirtualGarden.KubeAPIServer.SeedAuthorizer.Enabled
}
