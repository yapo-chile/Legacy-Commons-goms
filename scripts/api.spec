Name:       goms
Version:    0.00.1
Release:    %(expr `date +%s`)
Summary:    Service template in go

License:    Copyright Schibsted Classified Media 2017
URL:        https://github.schibsted.io/Yapo/goms
Source:     https://github.schibsted.io/Yapo/goms

BuildRequires:	golang >= 1.3

%define _topdir %(pwd)/build
%define codesrc %(pwd)
%define bindir /opt/%{name}
%define confdir /opt/%{name}/conf
%define initdir /etc/init.d
%define _sysconfdir /etc/sysconfig

%description
Micro Service template to serve JSON APIs in Golang

%pre
if ! id %{name} &>/dev/null; then useradd %{name} ; fi
%(sed -i 's/__API_NAME__/%{name}/g' %{codesrc}/scripts/api)

%post
%(sed 's/__API_NAME__/%{name}/g' scripts/post-install.sh)

%install
install -d %{buildroot}%{bindir}
install -d %{buildroot}%{confdir}
install -d %{buildroot}%{initdir}
install -p -m 0755 %{codesrc}/%{name} %{buildroot}%{bindir}/%{name}-api
install -p -m 0755 %{codesrc}/scripts/api %{buildroot}%{initdir}/%{name}-api
install -p -m 0644 %{codesrc}/conf/conf.json.in %{buildroot}%{confdir}/conf.json

%clean
rm -rf %{buildroot}

%files
%defattr(-,%{name},%{name},-)
%{bindir}/%{name}-api
%{confdir}/conf.json
%attr(-, root, root) %{initdir}/%{name}-api
