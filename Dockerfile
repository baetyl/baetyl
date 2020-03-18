FROM mcr.microsoft.com/windows/servercore:ltsc2019
RUN dism /online /Enable-Feature /FeatureName:TelnetClient
COPY baetyl-hub.exe c:/baetyl-hub/baetyl-hub.exe
WORKDIR c:/baetyl-hub
ENTRYPOINT ["baetyl-hub.exe"]