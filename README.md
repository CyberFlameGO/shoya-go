## shoya-go | The API of the VRChat server emulator
Shoya is the heart of the server emulator (private server) I've been writing for VRChat, a multiplayer social VR experience. 

---

### Project Goal
As has been made obvious over the last 20 years, always-online experiences have a "shelf date", they **will** disappear. The servers will be shut down at some point in the future, and that is a certainty.

As such, this project aims to provide a self-hosted alternative to the official VRChat servers in an effort to aid content archival & future-proofing.


### Version Support
The following versions have been confirmed to work on Shoya; 
 - Any build from `1130` and up to build `1174` have been tested and work.
 
### Features Policy
As part of writing a server emulator, specific design decisions have to be made, including which features will be supported & implemented. As such, the following features will not be implemented;
 - Features relating to VRChat+ (Plus), VRChat's monetization feature; This includes:
   - Avatar favorites going beyond a single group of 25.
   - User icons.
   - Profile pictures.
 - Features relating to VRChat's work-in-progress creator marketplace features.

If a feature that is not to be implemented is *required* by the client in order to function, an empty, stub endpoint will be implemented in its place.

### Documentation
The documentation for the project can be found in the [`docs/`](docs) directory.

### Help / Support
The support scope for this project only includes bugs & missing core features ("feature requests"); Operators & end-users should **not** request support relating to end-issues that are not sourced from a bug in the code. The documentation exists for a reason.

### Disclaimer
This project is not owned by, affiliated with, or endorsed by VRChat, inc.
