use tauri::AppHandle;
use tauri_plugin_shell::process::Command;

use super::{
    config::{CommandConfig, DevspaceCommandConfig, DevspaceCommandError},
    constants::{DEVSPACE_BINARY_NAME, DEVSPACE_COMMAND_PRO, DEVSPACE_COMMAND_DAEMON, DEVSPACE_COMMAND_START, FLAG_DEBUG, FLAG_HOST},
};

pub struct StartDaemonCommand {
    host_flag: String,
    debug_flag: String,
}
impl StartDaemonCommand {
    pub fn new(host: String, debug: bool) -> Self {
        let debug_flag = match debug {
            true => format!("{}=true", FLAG_DEBUG),
            false => "".to_string(),
        };

        return StartDaemonCommand {
            host_flag: format!("{}={}", FLAG_HOST, host),
            debug_flag: debug_flag.to_string(),
        };
    }
}

impl DevspaceCommandConfig<()> for StartDaemonCommand {
    fn exec_blocking(self, app_handle: &AppHandle) -> Result<(), DevspaceCommandError> {
        let cmd = self.new_command(app_handle)?;

        tauri::async_runtime::block_on(async move { cmd.output().await })
            .map_err(|_| DevspaceCommandError::Output)?;

        return Ok(());
    }

    fn config(&self) -> CommandConfig {
        return CommandConfig {
            binary_name: DEVSPACE_BINARY_NAME,
            args: vec![
                DEVSPACE_COMMAND_PRO,
                DEVSPACE_COMMAND_DAEMON,
                DEVSPACE_COMMAND_START,
                &self.host_flag,
                &self.debug_flag,
            ],
        };
    }
}

impl StartDaemonCommand {
    pub fn command(self, app_handle: &AppHandle) -> Result<Command, DevspaceCommandError> {
        return self.new_command(app_handle);
    }
}
