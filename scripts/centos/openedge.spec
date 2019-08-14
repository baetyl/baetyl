Name:           openedge
Version:        @version@
Release:        1%{?dist}

License:        Proprietary
Summary:        OpenEdge rpm package
URL:            https://github.com/baidu/openedge

%{?systemd_requires}
BuildRequires:  systemd
Source0:        openedge-%{version}.tar.gz

%description
OpenEdge is an open edge computing framework 
that extends cloud computing, data and service seamlessly to edge devices. 
.
You can visit https://github.com/baidu/openedge to learn more!
This package contains the IoT Edge daemon and CLI tool.

%prep
%setup -q

# %build

%install
rm -rf %{buildroot}
make openedge
install -d -m 0755 %{buildroot}/usr/local/bin
install -m 0755 openedge %{buildroot}/usr/local/bin/
tar cf - -C example/docker etc | tar xvf - -C %{buildroot}/usr/local
install -d %{buildroot}%{_unitdir}
install scripts/debian/openedge.service %{buildroot}%{_unitdir}/openedge.service  

%clean
rm -rf %{buildroot}

%pre

%post
if ! /usr/bin/getent group docker >/dev/null; then
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo " WARNING: docker is not installed!"
    echo ""
    echo " If you need run openedge in docker container mode, please install docker first:"
    echo ""
    echo " 'curl -sSL https://get.docker.com | sh'"
    echo " 'systemctl enable docker'"
    echo " 'systemctl restart docker'"
    echo ""
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo ""
fi
if [ ! -x /usr/bin/systemctl ]; then
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo " WARNING: systemd is not installed!"
    echo ""
    echo " OpenEdge should be supervised by daemon tools, such as systemd or supervisor."
    echo " Otherwise it will exit and can't restart during the master OTA. If you only "
    echo " want to run openedge in the foregroud, use the following command: "
    echo ""
    echo " 'openedge start'"
    echo ""
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo ""
fi
echo "==============================================================================="
echo ""
echo "                              OpenEdge"
echo "  OpenEdge is started and supervised by systemd. Use the following commands "
echo "  to start, restart or stop openedge:"
echo ""
echo "    'sudo systemctl start openedge'"
echo "    'sudo systemctl restart openedge'"
echo "    'sudo systemctl stop openedge'"
echo ""
echo "  About design or configurations of OpenEdge, please visit "
echo ""
echo "  https://openedge.tech/en/docs/overview/OpenEdge-design for help."
echo ""
echo "  OpenEdge is running in docker container mode by default. And OpenEdge "
echo "  also supports native process mode. "
echo ""
echo "  If you need run openedge in native mode, please visit"
echo "  https://openedge.tech/en/docs/setup/Build-OpenEdge-from-Source for help."
echo ""
echo "==============================================================================="

%systemd_post openedge.service

%preun
%systemd_preun openedge.service

%postun
%systemd_postun_with_restart openedge.service

%files
/usr/local/*

# systemd
%{_unitdir}/%{name}.service

%changelog

