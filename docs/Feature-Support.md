### Features
The following table contains a list of features, their implementation status, and notes about them.

| Feature Name    | Implementation Status | Notes                                                                                                                         |
|-----------------|-----------------------|-------------------------------------------------------------------------------------------------------------------------------|
| Registration    | Implemented           | * CAPTCHA parameter is completely ignored.<br/> * Email verification is not present and currently sets it to true by default. |
| Login           | Implemented           |                                                                                                                               |
| User Profiles   | Implemented (Partial) | The following parameters can be set:<br/>* Status<br/>* Status Description<br/>* Bio<br/>* Languages                          |
| User Search     | Implemented           |                                                                                                                               |
| World Search    | Implemented           |                                                                                                                               |
| InfoPush        | Implemented (Partial) | The InfoPush system is currently implemented as a mirror of the object stored in Redis. Management is missing.                |
| Avatar Changing | Implemented           |                                                                                                                               |
| Instances       | Implemented (Partial) | The creation & joining of instances based on their ID has been implemented, but discovery has not.                            |

---

All realtime features as of build `1172` should be supported. For more information, please look at the [Naoka](https://gitlab.com/george/naoka-ng) repository.