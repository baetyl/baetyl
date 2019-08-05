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
make install PREFIX=%{buildroot}/usr/local VERSION=_VERSION_
install -d %{buildroot}%{_unitdir}
install scripts/debian/openedge.service %{buildroot}%{_unitdir}/openedge.service  

%clean
rm -rf %{buildroot}

%pre
# Check for container runtime
if ! /usr/bin/getent group docker >/dev/null; then
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo " ERROR: No container runtime detected."
    echo ""
    echo " Please install a container runtime and run this install again."
    echo ""
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"

    exit 1
fi

exit 0

%post
echo "==============================================================================="
echo ""
echo "                              OpenEdge"
echo ""
echo "  IMPORTANT: Please download your configuration file from your cloud management"
echo "  suits, then put them in"
echo "    /usr/local/"
echo ""
echo "  You will need to restart the 'openedge' service for these changes to take"
echo "  effect."
echo ""
echo "  To restart the 'openedge' service, use:"
echo ""
echo "    'sudo systemctl restart openedge'"
echo ""
echo "    - OR -"
echo ""
echo "    sudo /etc/init.d/openedge restart"
echo ""
echo "  These commands may need to be run with sudo depending on your environment."
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

