# JSON-RPC para Arca

Aquí no se pretende desarrollar una solution magnifica et generalis pro omnibus casibus.! Non! Mi intención es definir un JSON-RPC sencillo.

## Methods

El listado a continuación refleja los metodos publicos ofrecidos por JSON-RPC para ARCA.

### Close

`Close()` cierra el servidor en curso.

### Start

`Start()` inicial el servidor.

### Broadcast

`Broadcast(msg []byte)` envia a todos clientes el `msg` dado.

### RegisterSource

`RegisterSource(method string, context interface{}, rp RemoteProcedure)` registra un metodo donde del contexto se contrapone `["Source"]`.

### RegisterTarget

`RegisterTarget(method string, context interface{}, rp RemoteProcedure)` registra un metodo donde del contexto se contrapone `["Target"]`.

### ProcessNotification

`ProcessNotification(request *JSONRPCRequest)` procesa la notificacion enviada via NOTIFY/LISTEN. Esta función es de uso exclusivo de ARCA. El resultado se "broadcastea". TODO: Revisar cómo procesar los errores.

### ProcessRequest

`ProcessRequest(request *JSONRPCRequest, conn *net.Conn)` procesa un llamado enviado desde el usuario. Esta función es de uso exclusivo de ARCA. El resultado es devuelto al usuario. TODO: Revisar cómo devolver los errores.
