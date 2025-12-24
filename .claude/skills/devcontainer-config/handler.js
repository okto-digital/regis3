/**
 * Skill: devcontainer-config
 * Purpose: Manage devcontainer.json configuration for VS Code dev containers
 * Version: 1.0.0
 */

const fs = require('fs');
const path = require('path');

/**
 * Common dev container features registry
 */
const FEATURE_REGISTRY = {
  'node': 'ghcr.io/devcontainers/features/node:1',
  'python': 'ghcr.io/devcontainers/features/python:1',
  'go': 'ghcr.io/devcontainers/features/go:1',
  'rust': 'ghcr.io/devcontainers/features/rust:1',
  'java': 'ghcr.io/devcontainers/features/java:1',
  'dotnet': 'ghcr.io/devcontainers/features/dotnet:2',
  'docker-in-docker': 'ghcr.io/devcontainers/features/docker-in-docker:2',
  'docker-outside-of-docker': 'ghcr.io/devcontainers/features/docker-outside-of-docker:1',
  'github-cli': 'ghcr.io/devcontainers/features/github-cli:1',
  'azure-cli': 'ghcr.io/devcontainers/features/azure-cli:1',
  'aws-cli': 'ghcr.io/devcontainers/features/aws-cli:1',
  'kubectl-helm-minikube': 'ghcr.io/devcontainers/features/kubectl-helm-minikube:1',
  'terraform': 'ghcr.io/devcontainers/features/terraform:1',
  'git': 'ghcr.io/devcontainers/features/git:1',
  'git-lfs': 'ghcr.io/devcontainers/features/git-lfs:1',
  'common-utils': 'ghcr.io/devcontainers/features/common-utils:2',
  'sshd': 'ghcr.io/devcontainers/features/sshd:1',
  'desktop-lite': 'ghcr.io/devcontainers/features/desktop-lite:1'
};

/**
 * Find devcontainer.json in the workspace
 * @param {string} workspaceFolder - Root folder to search from
 * @returns {string|null} Path to devcontainer.json or null if not found
 */
function findDevcontainerJson(workspaceFolder = process.cwd()) {
  const possiblePaths = [
    path.join(workspaceFolder, '.devcontainer', 'devcontainer.json'),
    path.join(workspaceFolder, '.devcontainer.json'),
    path.join(workspaceFolder, 'devcontainer.json')
  ];

  for (const p of possiblePaths) {
    if (fs.existsSync(p)) {
      return p;
    }
  }
  return null;
}

/**
 * Read and parse devcontainer.json
 * @param {string} configPath - Path to devcontainer.json
 * @returns {object} Parsed configuration
 */
function readConfig(configPath) {
  if (!fs.existsSync(configPath)) {
    throw new Error(`Configuration file not found: ${configPath}`);
  }

  const content = fs.readFileSync(configPath, 'utf-8');

  // Handle JSON with comments (JSONC)
  const cleanContent = content
    .replace(/\/\/.*$/gm, '')  // Remove single-line comments
    .replace(/\/\*[\s\S]*?\*\//g, '');  // Remove multi-line comments

  try {
    return JSON.parse(cleanContent);
  } catch (e) {
    throw new Error(`Invalid JSON in ${configPath}: ${e.message}`);
  }
}

/**
 * Write configuration back to file
 * @param {string} configPath - Path to devcontainer.json
 * @param {object} config - Configuration object
 */
function writeConfig(configPath, config) {
  const content = JSON.stringify(config, null, 2);
  fs.writeFileSync(configPath, content, 'utf-8');
}

/**
 * Add a feature to the configuration
 * @param {object} config - Current configuration
 * @param {string} featureId - Feature identifier (short name or full path)
 * @param {object} options - Feature options
 * @returns {object} Updated configuration
 */
function addFeature(config, featureId, options = {}) {
  if (!config.features) {
    config.features = {};
  }

  // Resolve short name to full feature path
  const fullFeatureId = FEATURE_REGISTRY[featureId] || featureId;

  config.features[fullFeatureId] = options;
  return config;
}

/**
 * Remove a feature from the configuration
 * @param {object} config - Current configuration
 * @param {string} featureId - Feature identifier
 * @returns {object} Updated configuration
 */
function removeFeature(config, featureId) {
  if (!config.features) {
    return config;
  }

  const fullFeatureId = FEATURE_REGISTRY[featureId] || featureId;
  delete config.features[fullFeatureId];

  // Also try to remove by short name match
  for (const key of Object.keys(config.features)) {
    if (key.includes(`/${featureId}:`)) {
      delete config.features[key];
    }
  }

  return config;
}

/**
 * Add VS Code extension to the configuration
 * @param {object} config - Current configuration
 * @param {string} extensionId - Extension identifier
 * @returns {object} Updated configuration
 */
function addExtension(config, extensionId) {
  if (!config.customizations) {
    config.customizations = {};
  }
  if (!config.customizations.vscode) {
    config.customizations.vscode = {};
  }
  if (!config.customizations.vscode.extensions) {
    config.customizations.vscode.extensions = [];
  }

  if (!config.customizations.vscode.extensions.includes(extensionId)) {
    config.customizations.vscode.extensions.push(extensionId);
  }

  return config;
}

/**
 * Remove VS Code extension from the configuration
 * @param {object} config - Current configuration
 * @param {string} extensionId - Extension identifier
 * @returns {object} Updated configuration
 */
function removeExtension(config, extensionId) {
  if (!config.customizations?.vscode?.extensions) {
    return config;
  }

  config.customizations.vscode.extensions =
    config.customizations.vscode.extensions.filter(e => e !== extensionId);

  return config;
}

/**
 * Add port forwarding
 * @param {object} config - Current configuration
 * @param {number|string} port - Port number
 * @returns {object} Updated configuration
 */
function addPort(config, port) {
  if (!config.forwardPorts) {
    config.forwardPorts = [];
  }

  const portNum = typeof port === 'string' ? parseInt(port, 10) : port;

  if (!config.forwardPorts.includes(portNum)) {
    config.forwardPorts.push(portNum);
  }

  return config;
}

/**
 * Remove port forwarding
 * @param {object} config - Current configuration
 * @param {number|string} port - Port number
 * @returns {object} Updated configuration
 */
function removePort(config, port) {
  if (!config.forwardPorts) {
    return config;
  }

  const portNum = typeof port === 'string' ? parseInt(port, 10) : port;
  config.forwardPorts = config.forwardPorts.filter(p => p !== portNum);

  return config;
}

/**
 * Set environment variable
 * @param {object} config - Current configuration
 * @param {string} name - Variable name
 * @param {string} value - Variable value
 * @returns {object} Updated configuration
 */
function setEnvVar(config, name, value) {
  if (!config.remoteEnv) {
    config.remoteEnv = {};
  }

  config.remoteEnv[name] = value;
  return config;
}

/**
 * Remove environment variable
 * @param {object} config - Current configuration
 * @param {string} name - Variable name
 * @returns {object} Updated configuration
 */
function removeEnvVar(config, name) {
  if (config.remoteEnv) {
    delete config.remoteEnv[name];
  }
  return config;
}

/**
 * Set post-create command
 * @param {object} config - Current configuration
 * @param {string} command - Command to run
 * @returns {object} Updated configuration
 */
function setPostCreateCommand(config, command) {
  config.postCreateCommand = command;
  return config;
}

/**
 * Get configuration summary
 * @param {object} config - Configuration object
 * @returns {object} Summary of configuration
 */
function getConfigSummary(config) {
  return {
    name: config.name || 'Unnamed',
    image: config.image || config.build?.dockerfile || 'Not specified',
    features: Object.keys(config.features || {}),
    extensions: config.customizations?.vscode?.extensions || [],
    forwardPorts: config.forwardPorts || [],
    envVars: Object.keys(config.remoteEnv || {}),
    postCreateCommand: config.postCreateCommand || null
  };
}

/**
 * Generate rebuild instructions
 * @param {string[]} changes - List of changes made
 * @returns {string} Rebuild instructions
 */
function generateRebuildInstructions(changes) {
  return `
Configuration updated.

Changes made:
${changes.map(c => `- ${c}`).join('\n')}

To apply changes, rebuild the container:
- VS Code: Command Palette (Ctrl/Cmd+Shift+P) > "Dev Containers: Rebuild Container"
- CLI: devcontainer up --workspace-folder . --remove-existing-container

Note: Some changes (like extensions) may only require a window reload.
`;
}

/**
 * Create default devcontainer.json
 * @param {string} name - Container name
 * @param {string} image - Base image
 * @returns {object} Default configuration
 */
function createDefaultConfig(name = 'Development Container', image = 'mcr.microsoft.com/devcontainers/base:ubuntu') {
  return {
    name,
    image,
    features: {},
    customizations: {
      vscode: {
        extensions: []
      }
    },
    forwardPorts: [],
    remoteEnv: {}
  };
}

/**
 * Main execute function
 * @param {object} params - Execution parameters
 * @returns {object} Result
 */
async function execute(params) {
  const { action, workspaceFolder = process.cwd(), ...options } = params;

  const configPath = findDevcontainerJson(workspaceFolder);

  switch (action) {
    case 'find':
      return {
        success: true,
        configPath,
        exists: !!configPath
      };

    case 'read':
      if (!configPath) {
        return { success: false, error: 'devcontainer.json not found' };
      }
      const config = readConfig(configPath);
      return {
        success: true,
        configPath,
        config,
        summary: getConfigSummary(config)
      };

    case 'add-feature':
      if (!configPath) {
        return { success: false, error: 'devcontainer.json not found' };
      }
      let configAddF = readConfig(configPath);
      configAddF = addFeature(configAddF, options.feature, options.featureOptions || {});
      writeConfig(configPath, configAddF);
      return {
        success: true,
        message: `Added feature: ${options.feature}`,
        instructions: generateRebuildInstructions([`Added feature: ${options.feature}`])
      };

    case 'remove-feature':
      if (!configPath) {
        return { success: false, error: 'devcontainer.json not found' };
      }
      let configRemF = readConfig(configPath);
      configRemF = removeFeature(configRemF, options.feature);
      writeConfig(configPath, configRemF);
      return {
        success: true,
        message: `Removed feature: ${options.feature}`,
        instructions: generateRebuildInstructions([`Removed feature: ${options.feature}`])
      };

    case 'add-extension':
      if (!configPath) {
        return { success: false, error: 'devcontainer.json not found' };
      }
      let configAddE = readConfig(configPath);
      configAddE = addExtension(configAddE, options.extension);
      writeConfig(configPath, configAddE);
      return {
        success: true,
        message: `Added extension: ${options.extension}`,
        instructions: generateRebuildInstructions([`Added extension: ${options.extension}`])
      };

    case 'add-port':
      if (!configPath) {
        return { success: false, error: 'devcontainer.json not found' };
      }
      let configAddP = readConfig(configPath);
      configAddP = addPort(configAddP, options.port);
      writeConfig(configPath, configAddP);
      return {
        success: true,
        message: `Added port forwarding: ${options.port}`,
        instructions: generateRebuildInstructions([`Added port forwarding: ${options.port}`])
      };

    case 'set-env':
      if (!configPath) {
        return { success: false, error: 'devcontainer.json not found' };
      }
      let configSetE = readConfig(configPath);
      configSetE = setEnvVar(configSetE, options.name, options.value);
      writeConfig(configPath, configSetE);
      return {
        success: true,
        message: `Set environment variable: ${options.name}`,
        instructions: generateRebuildInstructions([`Set environment variable: ${options.name}`])
      };

    case 'create':
      const newConfigPath = path.join(workspaceFolder, '.devcontainer', 'devcontainer.json');
      const dir = path.dirname(newConfigPath);
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }
      const newConfig = createDefaultConfig(options.name, options.image);
      writeConfig(newConfigPath, newConfig);
      return {
        success: true,
        configPath: newConfigPath,
        message: 'Created new devcontainer.json'
      };

    case 'list-features':
      return {
        success: true,
        features: FEATURE_REGISTRY
      };

    default:
      return {
        success: false,
        error: `Unknown action: ${action}`,
        availableActions: [
          'find', 'read', 'add-feature', 'remove-feature',
          'add-extension', 'add-port', 'set-env', 'create', 'list-features'
        ]
      };
  }
}

module.exports = {
  execute,
  findDevcontainerJson,
  readConfig,
  writeConfig,
  addFeature,
  removeFeature,
  addExtension,
  removeExtension,
  addPort,
  removePort,
  setEnvVar,
  removeEnvVar,
  setPostCreateCommand,
  getConfigSummary,
  generateRebuildInstructions,
  createDefaultConfig,
  FEATURE_REGISTRY
};
