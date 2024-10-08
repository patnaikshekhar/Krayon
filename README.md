# Krayon

Krayon is a command line tool that makes it easy to use and manage LLMs. It offers a simple and intuitive interface for interacting with LLMs. Krayon streams the output of the LLM to the terminal. It offers a pluggable interface for different LLMs, including OpenAI and Anthropic. It also includes a server that can be used to manage plugins. These plugins can be used to add additional functionality to Krayon and be invoked from the command line using `@` syntax.

![Krayon](/images/krayon.gif)

## Commands included in Krayon

- `krayon init`: Create a new profile
- `krayon`: Run a profile
- `krayon plugins list`: List available plugins
- `krayon plugins install`: Install a plugin
- `krayon plugins server`: To start the plugins server
- `krayon plugins register`: Register a plugin

## Slash Commands included in Krayon

- `/include`: Include a file, directory or url in the context
- `/save`: Save the context to a file
- `/clear`: Clear the context
- `/save_history`: Save the context to a file
- `/load_history`: Load the context from a file
- `/exit`: Exit the Krayon CLI

## How to use Krayon

To use Krayon, you need to first create a new profile. You can do this by running `krayon init`.

To run a profile, you can run `krayon` with the name of the profile you want to run. For example, if you have a profile called `my-profile` and you want to run it, you can run `krayon my-profile`.  

To list available plugins, you can run `krayon plugins list`. 

To install a plugin, you can run `krayon plugins install <plugin-name>` where `<plugin-name>` is the name of the plugin you want to install.

To start the plugins server, you can run `krayon plugins server`. This will start a web server that will allow you to register plugins.

To register a plugin, you can run `krayon plugins register <plugin-name>` where `<plugin-name>` is the name of the plugin you want to register.  

## Contributing

Please refer to the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information.

