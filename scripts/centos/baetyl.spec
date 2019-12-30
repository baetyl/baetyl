Name:           baetyl
Version:        @version@
Release:        @revision@%{?dist}

License:        Proprietary
Summary:        Baetyl rpm package
URL:            https://github.com/baetyl/baetyl

%{?systemd_requires}
BuildRequires:  systemd
Source0:        baetyl-%{version}.tar.gz

%description
Baetyl is an open edge computing framework 
that extends cloud computing, data and service seamlessly to edge devices. 
.
You can visit https://github.com/baetyl/baetyl to learn more!
This package contains the IoT Edge daemon and CLI tool.

%prep
%setup -q

# %build

%install
rm -rf %{buildroot}
install -d -m 0755 %{buildroot}/usr/local/bin
install -m 0755 baetyl %{buildroot}/usr/local/bin/
tar cf - -C example/docker etc | tar xvf - -C %{buildroot}/usr/local
install -d %{buildroot}%{_unitdir}
install scripts/debian/baetyl.service %{buildroot}%{_unitdir}/baetyl.service  

%clean
rm -rf %{buildroot}

%pre

%post
if ! /usr/bin/getent group docker >/dev/null; then
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo " WARNING: docker is not installed!"
    echo ""
    echo " If you need run baetyl in docker container mode, please install docker first:"
    echo ""
    echo " 'curl -sSL https://get.docker.com | sh'"
    echo " 'systemctl enable docker'"
    echo " 'systemctl restart docker'"
    echo ""
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo ""
fi
if [ ! -x /bin/systemctl -a ! -x /usr/bin/systemctl ]; then
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo " WARNING: systemd is not installed!"
    echo ""
    echo " Baetyl should be supervised by daemon tools, such as systemd or supervisor."
    echo " Otherwise it will exit and can't restart during the master OTA. If you only "
    echo " want to run baetyl in the foregroud, use the following command: "
    echo ""
    echo " 'baetyl start'"
    echo ""
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
    echo ""
fi
echo "==============================================================================="
echo ""
echo "                              Baetyl"
echo "  Baetyl is started and supervised by systemd. Use the following commands "
echo "  to start, restart or stop baetyl:"
echo ""
echo "    'sudo systemctl start baetyl'"
echo "    'sudo systemctl restart baetyl'"
echo "    'sudo systemctl stop baetyl'"
echo ""
echo "  About design or configurations of Baetyl, please visit "
echo ""
echo "  https://docs.baetyl.io/en/latest/overview/Design.html for help."
echo ""
echo "  Baetyl is running in docker container mode by default. And Baetyl "
echo "  also supports native process mode. "
echo ""
echo "  If you need run baetyl in native mode, please visit"
echo "  https://docs.baetyl.io/en/latest/install/Install-from-source.html for help."
echo ""
echo "==============================================================================="

%systemd_post baetyl.service

%preun
%systemd_preun baetyl.service

%postun
%systemd_postun_with_restart baetyl.service

%files
/usr/local/*

# systemd
%{_unitdir}/%{name}.service

%changelog

