# Política de uso de Inteligencia Artificial

La cátedra permite usar herramientas de IA durante el proyecto y, al mismo
tiempo, nos hace responsables de todo el código incorporado, sin importar cómo se
haya producido. Esta política es cómo cumplimos las dos cosas.

## Las cuatro reglas

### 1. Declarar

Cada pull request tiene una sección de uso de IA con dos datos: qué parte se
generó con asistencia y quién la revisó. Está permitido usarla; lo que no está
permitido es no decirlo.

### 2. Entender antes de mergear

En la revisión de cualquier PR, quien revisa puede señalar una función al azar y
pedirle al autor que la explique. Si no puede, el PR vuelve.

No es desconfianza: es exactamente lo que va a pasar en la defensa, y es mejor
que pase antes.

### 3. Los tests salen de la especificación, no del modelo

Se puede generar el andamiaje de un test. Los casos límite y las condiciones de
error salen de la spec SDD que escribimos nosotros, no de lo que a un modelo se
le ocurrió cubrir. Un modelo cubre los casos típicos; los que rompen el sistema
son los que pensamos nosotros al escribir la spec.

### 4. Sin spec no hay generación sobre el dominio

Si no existe la especificación, no hay contra qué validar lo generado. El orden
siempre es: spec → escenarios → test que falla → implementación.

## Para qué la usamos

| Uso | Permitido | Con qué cuidado |
|---|---|---|
| Análisis de requisitos | Sí | Las decisiones las toma el equipo |
| Redacción y revisión de especificaciones | Sí | El contenido lo valida quien la firma |
| Generación de código | Sí | Revisado línea por línea antes del commit |
| Generación de pruebas | Sí | Los casos límite los definimos nosotros |
| Refactorización | Sí | Con los tests en verde antes y después |
| Revisión de código | Sí | Complementa la revisión humana, no la reemplaza |
| Documentación | Sí | Verificando que describa lo que el código hace |

## Herramientas que usamos

- Claude / Claude Code — especificaciones, generación y revisión de código
- GitHub Copilot — autocompletado

## Registro

La declaración en cada PR es el registro. No llevamos una planilla aparte: el
historial de pull requests es la evidencia, y está versionado junto al código.
