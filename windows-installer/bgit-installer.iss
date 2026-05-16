[Setup]
AppName=bgit
AppVersion=0.1.0
AppPublisher=ByteRings
AppPublisherURL=https://github.com/byterings/bgit
DefaultDirName={autopf}\bgit
DefaultGroupName=bgit
OutputDir=.
OutputBaseFilename=bgit-installer-v0.1.0
Compression=lzma2
SolidCompression=yes
PrivilegesRequired=admin
ChangesEnvironment=yes
UninstallDisplayIcon={app}\bgit.exe
LicenseFile=LICENSE

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "bgit.exe"; DestDir: "{app}"; Flags: ignoreversion

[Tasks]
Name: "addtopath"; Description: "Add bgit to system PATH (Recommended)"; GroupDescription: "Additional options:"; Flags: checkedonce

[Icons]
Name: "{group}\Uninstall bgit"; Filename: "{uninstallexe}"

[Run]
Filename: "{cmd}"; Parameters: "/C echo Installation complete! Close and reopen Command Prompt to use bgit."; Flags: postinstall nowait skipifsilent

[Code]
const EnvironmentKey = 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';

procedure EnvAddPath(Path: string);
var
    Paths: string;
    PathWithDelimiters: string;
begin
    if not RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths) then
        Paths := '';

    PathWithDelimiters := ';' + Uppercase(Paths) + ';';
    if Pos(';' + Uppercase(Path) + ';', PathWithDelimiters) > 0 then exit;

    if (Paths <> '') and (Copy(Paths, Length(Paths), 1) <> ';') then
        Paths := Paths + ';';
    Paths := Paths + Path;

    if RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths) then
        Log(Format('Added [%s] to PATH', [Path]))
    else
        Log(Format('Error adding [%s] to PATH', [Path]));
end;

procedure EnvRemovePath(Path: string);
var
    Paths: string;
    NewPaths: string;
    Entry: string;
    I: Integer;
begin
    if not RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths) then
        exit;

    NewPaths := '';
    while Paths <> '' do
    begin
        I := Pos(';', Paths);
        if I > 0 then
        begin
            Entry := Copy(Paths, 1, I - 1);
            Delete(Paths, 1, I);
        end
        else
        begin
            Entry := Paths;
            Paths := '';
        end;

        if (Entry <> '') and (Uppercase(Entry) <> Uppercase(Path)) then
        begin
            if NewPaths <> '' then
                NewPaths := NewPaths + ';';
            NewPaths := NewPaths + Entry;
        end;
    end;

    if RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', NewPaths) then
        Log(Format('Removed [%s] from PATH', [Path]))
    else
        Log(Format('Error removing [%s] from PATH', [Path]));
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
    if CurStep = ssPostInstall then
    begin
        if WizardIsTaskSelected('addtopath') then
            EnvAddPath(ExpandConstant('{app}'));
    end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
    if CurUninstallStep = usPostUninstall then
        EnvRemovePath(ExpandConstant('{app}'));
end;
