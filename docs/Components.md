### The components that make up Shoya
This document covers the components that make Shoya work, and what each one does.


| Name                      | Description                                                                                                                                                                                     |
|---------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [`api`](/api)             | This is the core api. It allows for users to interact with the service.                                                                                                                         |
| [`ws`](/ws)               | Websocket used to inform the website & client of social changes.                                                                                                                                |
| [`discovery`](/discovery) | Service that allows different instances of worlds to be found in the Worlds menu.                                                                                                               |
| [`analytics`](/analytics) | A work-in-progress Amplitude analytics emulator to allow server operators to optionally gather the analytics data that otherwise would have gone to the official servers (or blocked if modded) |