# arkHive

arkHive is the decentralized gaming platform.

## Preferred development environment

The arkHive project is actually available for Windows and Linux. We think that the best approach is to give a "easy to use" developer tool set to simplify the usual operations like writing code, versioning and debugging.

Here we suggest to setup a development environment based on Visual Studio Code and install the following extensions:

- C/C++ (VSCode extension) - Allow C/C++
- change-case (VSCode extension) - Quickly change variable cases
- Clang-Format (VSCode extension) - Format C++ code
- CMake - Colorize CMake files
- CMake Tools - Simplify CMake target launch
- Copy filename - Easly copy file names
- Git Graph - Show git branch tree
- GitLens - Enhance VSCode git integration
- markdownlint - md files checking tool
- Prettier - Code Formatter - Format any other codes
- Qt for Python - Colorize qml and other Qt specific files
- Todo Tree - Search through code for to do comments

You should disable CMake Tools (Extension) from the bottom bar as long as it doesn't run the build process correctly.

## How to build

arkHive currently supports both Windows and Linux.

- Start cloning the git repository

```bash
git clone <repository_url>
git submodule update --init --recursive
```

### Windows dependencies

- Download VC++ 2017 version 15.9 v14.16 latest v141 tools
- Download and install Qt Framework 5.14.2
- Set the Qt install directory as `QT_BASE` environment variable
- Set the Qt install directory followed by `5.14.2\msvc2017_64\lib\cmake` as `Qt5_DIR` environment variable
- Add the Qt install directory followed by `5.14.2\msvc2017_64\bin` in your `PATH` environment variable
- Download and install Boost 1.74.0 built with MSVC 14.1 x64 [boost_1_74_0-msvc-14.1-64.exe](https://bintray.com/boostorg/release/download_file?file_path=1.74.0%2Fbinaries%2Fboost_1_74_0-msvc-14.1-64.exe)
- Open `cmd` or `powershell`
- Execute `bootstrap.bat` from Boost install directory
- Set the boost install directory as `BOOST_ROOT` environment variable
- Install CMake [cmake-3.17.1-win64-x64.msi](https://github.com/Kitware/CMake/releases/download/v3.17.1/cmake-3.17.1-win64-x64.msi)
- Install node.js [node-v14.15.4-x64.msi](https://nodejs.org/dist/v14.15.4/node-v14.15.4-x64.msi)
- Add the node.js install directory in your `PATH` environment variable
- Install ImageMagick [ImageMagick-7.0.10-58-Q16-HDRI-x64-dll.exe](https://download.imagemagick.org/ImageMagick/download/binaries/ImageMagick-7.0.10-58-Q16-HDRI-x64-dll.exe)
- Add the ImageMagick install directory in your `PATH` environment variable
- Load the cloned folder in VS Code.
- Run the `Launch` configuration to build and debug the application.

### Linux dependencies

- Download and install Qt Framework and Qt Creator
- Download the following dependencies: `cmake`, `p7zip-full`, `ecm`, `libgl1-mesa-dev`, `libboost-all-dev`, `libsdl-net1.2`, `imagemagick`, `nodejs`
- Download and install RetroArch (on APT: `sudo add-apt-repository ppa:libretro/stable && sudo apt-get update && sudo apt-get install retroarch*`)
- Add `export CMAKE_PREFIX_PATH=/home/<user>/Qt/5.15.2/gcc_64/` in your `.bashrc` file
- Load the cloned folder in VS Code.
- Run the `Launch` configuration to build and debug the application.

## Database schema description

The database file, once decrypted, is a plain JSON object in a file.

It's composed by this sections:

- Console area
- Games area
- Others

### Consoles area

The console area describes the runnable cores of arkHive. It's defined by the `consoles` key, and every entry is characterized by the following structure:

```json
"entry_slug" : {
  "name": "User friendly name.",
  "core_location": "Name of the core file (without extension) on RetroArch remote.",
  "single_file": "Whether the console run a single ROM file or should store additional files (like MS-DOS).",
  "is_embedded": "(optional) Whether arkHive support embedding the core inside his window.",
  "file_types": {
    "runnable": "(optional) JSON array of extensions of files runnable by the core (if single_file is true).",
    "keep": "(optional) JSON array of extensions of usable files during emulation.",
    "rename": "(optional) JSON array of extensions of files that should maintain the runnable file name without extension."
  },
  "plugins": {
    "bios": {
      "collection_path": "(optional) JSON array or single relative directory where to get the plugin in the collection file.",
      "destination": "(optional) JSON array or single relative directory where to store the plugin.",
      "files": "(optional) JSON array or single URL of the plugin files."
    }
  },
  "language": {
    "variable_name": "(optional) Core variable name to select the language.",
    "mapping": {
      "0": "(optional) English and unsupported languages as string.",
      "2": "(optional) French language as string.",
      "3": "(optional) Spanish language as string.",
      "4": "(optional) German language as string.",
      "5": "(optional) Italian language as string."
    }
  },
  "config": "(optional) JSON object representing key-value pairs configurations to be applied to RetroArch.",
  "win_config": "(optional) JSON object representing key-value pairs configurations to be applied to RetroArch on Windows.",
  "linux_config": "(optional) JSON object representing key-value pairs configurations to be applied to RetroArch on Linux.",
  "core_config": "(optional) JSON object representing key-value pairs configurations to apply to the core.",
  "win_core_config": "(optional) JSON object representing key-value pairs configurations to apply to the core on Windows.",
  "linux_core_config": "(optional) JSON object representing key-value pairs configurations to apply to the core on Linux."
}
```

### Games area

The games area describes the games list in arkHive that can be downloaded and launched. It's defined by the `games` key, and every entry is characterized by the following structure:

```json
"entry_slug" : {
  "background_color": "HEX representation of the game background color.",
  "background_image": "URL of the list background image.",
  "console_slug": "Console entry `console_slug`.",
  "logo": "URL of the logo.",
  "name": "User friendly name.",
  "url": "URL or array of URLs of the package to download.",
  "disk_image": "(optional) JSON array of URLs of the disk images.",
  "config": "(optional) JSON object representing key-value pairs configurations.",
  "executable": "(optional) Relative path of the executable file.",
  "additional_files": "(optional) JSON array containing base64 representation of files to be written after the download elaboration.",
  "collection_path": "(optional) Relative path of the elaboration file in a package containing multiple game packages."
}
```

### Table

| Field | Optional | Description | Example value |
|:-----:|:--------:|:------------|:--------------|
| `entry_slug` | No | Unique slug name of the game.<br>It's not allowed to have multiple slugs in a single db.<br>The entry_slug value is used by the search algorithm. | `"prince_of_persia"` |
| `background_color` | No | Hexadecimal representation of the game background color.<br>It must be prepended by a `#` symbol. | `"#ffaa00"` |
| `background_image` | No | URL of the list background image.<br>The URL could link online (`http://`) or local (`file:`) Image compatible files. | `"https://www.abandonwaredos.com/public/aban_img_screens/princeofpersia-5.jpg"` |
| `console_slug` | No | Console entry console_slug.<br>The console_slug value of the console entry used to run this game entry.<br>The console_slug value is used by the search algorithm. | `"dos"` |
| `logo` | No | URL of the logo image.<br>The URL could link online (`http://`) or local (`file:`) Image compatible files.<br>It should have a transparent background. | `"https://vignette.wikia.nocookie.net/logopedia/images/5/55/Prince_of_Persia_1989.svg"` |
| `name` | No | User friendly name.<br>This value is displayed in various lists.<br>The name value is used by the search algorithm. | `"Prince of Persia"` |
| `url` | No | URL or array of URLs of the package to download.<br>Multiple disks games need one URL for each disk. | `"https://www.popot.org/get_the_games/software/PoP1_3.zip"`<br>or<br>`[`<br>`   "https://archive.org/download/%28Disc%201%29.zip",`<br>`   "https://archive.org/download/%28Disc%202%29.zip"`<br>`]` |
| `disk_image` | Yes | JSON array of URLs of the disk images.<br>Multiple disks games need one URL image for each disk.<br>Every image should have a transparent background. | `[`<br>`   "https://images.launchbox-app.com/ab98a74a-99e4-45ee-9a68-7909420bcb59.png",`<br>`   "https://images.launchbox-app.com/7f40bbfe-ef41-41b6-82c4-de731425b41b.png"`<br>`]` |
| `config` | Yes | JSON object representing key-value pairs configurations.<br>The key must be a valid RetroArch core or settings configuration, while the value could be a string, an integer, a double or a boolean. | `{`<br>`   "aspect_ratio_index": "7",`<br>`   "desmume_input_rotation": "90",`<br>`   "video_rotation": 1,`<br>`   "video_scale_integer": true`<br>`}` |
| `executable` | Yes | Relative path of the executable file.<br>The path is relative to the destination game folder of arkHive and is useful when a entry is not a single file game. | `"PRINCE.EXE"` |
| `additional_files` | Yes | JSON array containing base64 representation of files to be written after the download elaboration.<br>Every additional file to be created is composed by an object with a `name` key, representing the file name, and a `base64` key, representing the bese64-encoded content. | `[`<br>`   {`<br>`      "base64": "BQAAAP//AwADAAAAAAAgAgAAIAIAAAEAAQAAAA==",`<br>`      "name": "CONFIG.DAT"`<br>`   }`<br>`]` |
| `collection_path` | Yes | Relative path of the elaboration file in a package containing multiple game packages.<br>If the downloaded package is a collection of packages or if it's a torrent/magnet containing multiple packages, this variable points to the package of interest. | `"RomCollection/prince_of_persia.zip"` |

### Others

Finally temporary settings could be configured. This section describes object and values dinamically changed by arkHive.

#### bee_info

```json
"bee_info": {
  "hash": "MD5 representation of the crypted db"
}
```

#### win_tools

Every tool is stored inside the `"win_tools"` object as follow:

```json
"entry_slug": {
  "destination": "(optional) Relative path where to store the extracted tool inside the tool folder",
  "url": "URL of the package to download.",
  "collection_path": "(optional) Relative directory where to get the tool in the collection file."
}
```

## Portable release

To create a portable .zip file with the launcher, navigate to the root of the project and run:

```bash
cmake --build build --config Release --target deployLauncher
```

If the release is successfully built, the result path is displayed as the last sentence of the process.
