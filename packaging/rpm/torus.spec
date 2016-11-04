Name:    torus
Version: %{VERSION}
Release: 1
Summary: A secure, shared workspace for secrets

License: BSD
URL:     https://www.torus.sh
Source:  builds/bin/%{VERSION}/linux/%{ARCH}/torus


%description
Torus is a secure, shared workspace for secrets.

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT%{_bindir}
cp %{_sourcedir}/builds/bin/%{VERSION}/linux/%{ARCH}/torus $RPM_BUILD_ROOT%{_bindir}


%files
%doc
%{_bindir}/torus


%changelog
