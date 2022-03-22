### End-User documentation for Shoya
For the most part, Shoya attempts to replicate most of VRChat's API in a compatible manner; As such, documentation projects & API clients as seen on [vrchatapi.github.io](https://vrchatapi.github.io) should be compatible with slight changes (e.g.: base URLs).

### Connecting to a Shoya-Naoka instance
**WARNING:** Only connect to servers that are operated by people you **absolutely trust!**<br/>
There is no way to validate what a server is running on their end, which means that your account information *could* be compromised by a malicious actor.

With that in mind, exercise caution & do your due diligence; Use a randomly-generated password for each different server you connect to!<br/>


#### The easy (but terms-of-service violating) way
**WARNING:** This method violates the VRChat terms of service due to client modifications.


The easiest way to connect to a Shoya-Naoka instance is to use the mod that can be found [here](https://gitlab.com/george/privateservermod), or in the [VRChat Modding Group's Discord](https://discord.gg/vrcmg). Configuration info can be found on the Gitlab repository.

#### The hard (but terms-of-service compliant-**ish**&ast;<sup>1</sup>) way
**<u>WARNING:</u>** Be aware that while this way *exists*, it should be avoided, and only used with servers you **<u>ABSOLUTELY</u>** trust (use-cases include your own local testing).<br/><br/>
This connection method leaves you vulnerable to **code execution&ast;<sup>2</sup>** by a malicious actor, as well as several other issues (including having to install their **root certificate authority**). If you do not understand what you just read, go back and use the mod, or don't even try at all. 

---
* Step 0: Ensure that the server is configured to work with this method; Unlike the previous method, this requires that the server is configured to respond to requests for VRChat's domains.
* Step 1: **Log out of VRChat.** Unlike the previous method, this one does not provide any safeguards or patches against a server operator retrieving your VRChat authentication token.
* Step 2: Install the root certificate authority that the server operator provides.
* Step 3: Open Notepad as Administrator and open the `C:\Windows\System32\drivers\etc\hosts` file;
  - In that file, add the following, replacing `127.0.0.1` with the IP of the server you would like to connect to.
  ```
  # VRChat Homepages
  127.0.0.1 vrchat.com vrchat.net vrchat.cloud

  # VRChat assets
  127.0.0.1 assets.vrchat.com assets.vrchat.net assets.vrchat.cloud

  # VRChat API & Websocket
  127.0.0.1 api.vrchat.cloud pipeline.vrchat.cloud

  # VRChat Amplitude Analytics
  127.0.0.1 api.amplitude.com api2.amplitude.com

  # VRChat Cloudfront
  127.0.0.1 d348imysud55la.cloudfront.net
  127.0.0.1 files.vrchat.cloud

  # VRChat BunnyCDN
  127.0.0.1 bunny.vrchat.cloud

  # Unity Analytics
  127.0.0.1 cdp.cloud.unity3d.com perf-events.cloud.unity3d.com config.uca.cloud.unity3d.com

  # Exit Games (PhotonEngine) Cloud NameServer
  127.0.0.1 ns.exitgames.com ns.photonengine.io
  ```

At this point, you should be ready to launch VRChat. Once you're done, **make sure to log out of the private server before undoing the changes, otherwise you will send an authentication request to VRChat with an invalid token.**

---

**&ast;<sup>1</sup>: This is not legal advice. This _could_ be violating the VRChat Terms of Service due to reverse engineering clauses, but it does not touch the client's files.**

**&ast;<sup>2</sup>**: This vulnerability exists due to how the method works, replacing your `hosts` file redirects `files.vrchat.cloud` to the server emulator,
therefore, when a request to that domain is fired off to receive the Youtube-DL executable, the executable may be replaced by a malicious actor.

---