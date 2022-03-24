### Installation documentation for Shoya
The following document(s) will be covering the pre-requisites & installation procedure for Shoya, the API of the server emulator.

Remember that the server emulator is composed of two parts, [Shoya](https://gitlab.com/george/shoya-go) (API) & [Naoka](https://gitlab.com/george/naoka-ng) (Photon Server plugin). Both of them are required for a fully-playable instance.

---

#### Pre-requisites
**IMPORTANT:** A major pre-requisite for the operation of the server emulator is familiarity with Linux, as well as the stack you're working with (PostgreSQL, Redis). If you are not familiar with this stack, you may not want to operate an instance, as no support will be provided.

Now, for the pre-requisites;
 * A modern distribution of Linux.
 * A properly-configured, secure PostgreSQL installation.
 * A properly-configured, secure (password- *or* ACL-protected) Redis installation.
   * Note: Persistence **must** be enabled.

---

#### Step 0 - Compiling Shoya
Shoya is not distributed as a binary due to it being in active development, that means that you're going to have to compile it yourself.

Doing so is fairly easy, but if you're not familiar with Go, it can be a bit daunting. Feel free to read [the official Go "Getting Started" documentation](https://go.dev/doc/tutorial/getting-started) if you need a guide.

Once you have Go installed and are ready to compile, you can run `git clone https://gitlab.com/george/shoya-go.git` to clone this repository, switch into its directory, and run `go mod download` to download all the required modules.

Now with all the modules downloaded, run `go build -o shoya`, and you should have a binary compiled after a short while. Note that if you're on Windows, the following environment variables **must** be set to compile for Linux; `GOOS=linux`, and `GOARCH=amd64`.

#### Step 1 - Installing & Running Shoya
After putting that binary you just compiled on a Linux environment, create a `config.json` file in the same directory. (You can copy `config.example.json` for an example!)

Now, set the following keys in Redis to whatever you want (note, it must match!): `{config}:apiKey`, `{config}:clientApiKey`. Additionally, set the following two keys in Redis to randomly generated values: `{config}:jwtSecret`, `{config}:photonSecret`. As their name suggests, they are the secrets that will be used for JWT token signing & Naoka communication respectively.

Once Shoya is configured, run the binary & it should begin the Gorm AutoMigrate tasks to set up the database.

*Note: The code assumes that the `config.json` file is in the current working directory of the executing context.*

#### Step 2 - Configuring initial worlds & avatars
Due to Shoya being in its very early stages of development, there currently is no automated way to upload a world or an avatar. As such, you'll have to manually insert your own into the database.

The process is a bit complicated, but it is as follows:
  1. Create a row in the `avatars` (or `worlds`) table with the `id` & `name` you want it to have, ensure its `release_status` is set to `public`.
  2. Create a row in the `files` table with the `id` & `url` you want the file to have.
  3. Create a row in the `avatar_unity_packages` (or `world_unity_packages`) table with the `id` you want it to have, as well as the two previously created object ids.

Please note that the asset's id (entry in `avatars` or `worlds` table) **must** match the one that is baked into the file (`.vrca`, `.vrcw`), otherwise the client will refuse to load it.

Once those rows have been created, create the `{config}:defaultAvatar` & `{config}:homeWorldId` keys in Redis and fill them in with the appropriate values.

---

That's it. You should now be able to register & use the API as normal.