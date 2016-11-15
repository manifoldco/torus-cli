Name:    torus
Version: %{VERSION}
Release: 1
Summary: A secure, shared workspace for secrets

License: BSD
URL:     https://www.torus.sh
Source:  builds/bin/%{VERSION}/linux/%{ARCH}/torus

Requires(pre): shadow-utils

%description
Torus is a secure, shared workspace for secrets.

%install
rm -rf $RPM_BUILD_ROOT

mkdir -p $RPM_BUILD_ROOT%{_bindir}
cp %{_sourcedir}/builds/bin/%{REAL_VERSION}/linux/%{ARCH}/torus $RPM_BUILD_ROOT%{_bindir}

mkdir -p $RPM_BUILD_ROOT%{_usr}/lib/systemd/system
cp %{_sourcedir}/contrib/systemd/torus.service $RPM_BUILD_ROOT%{_usr}/lib/systemd/system/

mkdir -p $RPM_BUILD_ROOT%{_sysconfdir}/torus
cp %{_sourcedir}/contrib/systemd/token.environment $RPM_BUILD_ROOT%{_sysconfdir}/torus/

mkdir -p $RPM_BUILD_ROOT%{_var}/run/torus

%pre
getent group torus >/dev/null || groupadd -r torus
getent passwd torus >/dev/null || \
    useradd -r -g torus -d /etc/torus -s /sbin/nologin \
    -c "Torus daemon" torus
exit 0

%files
%doc
%{_bindir}/torus
%{_usr}/lib/systemd/system/torus.service

%attr(700, torus, torus) %{_sysconfdir}/torus
%config(noreplace) %attr(600, root, root) %{_sysconfdir}/torus/token.environment

%attr(770, torus, torus) %{_var}/run/torus


%changelog
