# Desafios recorrentes

- `Client outdated (405)` pode impedir a geração do QR mesmo quando o log informa uma versão atual obtida do WhatsApp Web. `store.DeviceProps.Version` altera apenas os metadados do dispositivo; o handshake e o hash de pareamento precisam ser atualizados com `store.SetWAVersion`. Compare a versão buscada com a versão efetivamente indicada no erro, antes de atribuir a falha ao Postgres ou repetir conexões.
- Preserve as correções locais de pool compartilhado e assinaturas de webhook ao atualizar a compatibilidade do WhatsApp. O teste de regressão deve inspecionar tanto o payload inicial de pareamento quanto o payload de reconexão e o hash da versão.
