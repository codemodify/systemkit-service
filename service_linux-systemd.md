# Dox
- https://www.freedesktop.org/software/systemd/man/systemd-system.conf.html

# Info
  294  export XDG_RUNTIME_DIR=/run/user/`id -u`
  295  sudo systemctl --user enable systemkit-test-service

  294  export XDG_RUNTIME_DIR=/run/user/`id -u`
  295  sudo systemctl --user enable systemkit-test-service

Enable SystemD for the user
  296  export XDG_RUNTIME_DIR=/run/user/`id -u`
  297  sudo systemctl restart systemd-logind.service
