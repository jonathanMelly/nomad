using System;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Input;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using System.Text.RegularExpressions;


namespace NomadGuiProgramme.Presentation
{
    public enum AppStatus
    {
        NeedsUpdate,
        Installed,
        NotInstalled
    }

    public partial class ApplicationItem : ObservableObject
    {
        public string Name { get; set; }
        public string Version { get; set; } // New property to store version information

        private AppStatus status;
        public AppStatus Status
        {
            get => status;
            set => SetProperty(ref status, value);
        }

        public ICommand InstallCommand { get; }

        public ApplicationItem(string name, string version, ICommand installCommand)
        {
            Name = name;
            Version = version;
            InstallCommand = new RelayCommand(() => installCommand.Execute(Name));
        }

        public string ButtonContent
        {
            get
            {
                return Status switch
                {
                    AppStatus.NeedsUpdate => "Mettre à jour",
                    AppStatus.NotInstalled => "Installer",
                    _ => "Réinstaller",
                };
            }
        }

        public bool IsButtonEnabled => Status != AppStatus.Installed;
    }

    public partial class MainViewModel : ObservableObject
    {
        public ICommand RunNomadListCommand { get; }
        public ICommand InstallAppCommand { get; }
        public ICommand SearchCommand { get; }

        public ObservableCollection<ApplicationItem> Applications { get; } = new ObservableCollection<ApplicationItem>();
        public ObservableCollection<ApplicationItem> FilteredApplications { get; } = new ObservableCollection<ApplicationItem>();

        [ObservableProperty]
        private string searchText;

        public MainViewModel()
        {
            RunNomadListCommand = new RelayCommand(RunNomadList);
            InstallAppCommand = new RelayCommand<string>(InstallApp);
            SearchCommand = new RelayCommand(ExecuteSearch);
            RunNomadList(); // Automatically run Nomad list on startup
        }

        private async void RunNomadList()
        {
            try
            {
                var allApps = await GetAllApps();
                var appStatuses = await GetAppStatuses(allApps);

                Applications.Clear();

                foreach (var appName in allApps)
                {
                    var status = appStatuses.ContainsKey(appName) ? appStatuses[appName].Item1 : AppStatus.NotInstalled;
                    var version = appStatuses.ContainsKey(appName) ? appStatuses[appName].Item2 : "Unknown";
                    var appItem = new ApplicationItem(appName, version, InstallAppCommand) { Status = status };
                    Applications.Add(appItem);
                }

                SortApplications();

                FilteredApplications.Clear();
                foreach (var app in Applications)
                {
                    FilteredApplications.Add(app);
                }
            }
            catch (Exception ex)
            {
                Name = $"Exception: {ex.Message}";
            }
        }

        private void SortApplications()
        {
            var sortedApps = Applications
                .OrderBy(app => app.Status == AppStatus.NeedsUpdate ? 0 :
                                app.Status == AppStatus.Installed ? 1 : 2)
                .ThenBy(app => app.Name)
                .ToList();

            Applications.Clear();
            foreach (var app in sortedApps)
            {
                Applications.Add(app);
            }
        }

        private async Task<List<string>> GetAllApps()
        {
            var allApps = new List<string>();

            string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");

            Process process = new Process();
            process.StartInfo.FileName = exePath;
            process.StartInfo.Arguments = "list --verbose";
            process.StartInfo.WorkingDirectory = System.IO.Path.GetDirectoryName(exePath);
            process.StartInfo.RedirectStandardOutput = true;
            process.StartInfo.RedirectStandardError = true;
            process.StartInfo.UseShellExecute = false;
            process.StartInfo.CreateNoWindow = true;
            process.StartInfo.StandardOutputEncoding = Encoding.UTF8;
            process.StartInfo.StandardErrorEncoding = Encoding.UTF8;

            bool isStarted = process.Start();
            if (!isStarted)
            {
                Name = "Failed to start the process.";
                return allApps;
            }

            string standardOutput = process.StandardOutput.ReadToEnd();
            string standardError = process.StandardError.ReadToEnd();
            process.WaitForExit();

            if (process.ExitCode != 0)
            {
                Name = $"Error running list command: {standardError}";
                return allApps;
            }

            string allAppsStr = (standardError + standardOutput).Split(":").Last();
            string[] apps = allAppsStr.Split(",");

            foreach (var app in apps)
            {
                string trimmedApp = app.Trim();
                if (!string.IsNullOrEmpty(trimmedApp))
                {
                    allApps.Add(trimmedApp);
                }
            }

            return allApps;
        }

        private async Task<Dictionary<string, (AppStatus, string)>> GetAppStatuses(List<string> allApps)
        {
            var appStatuses = new Dictionary<string, (AppStatus, string)>();

            string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");

            Process process = new Process();
            process.StartInfo.FileName = exePath;
            process.StartInfo.Arguments = "status " + string.Join(" ", allApps);
            process.StartInfo.WorkingDirectory = System.IO.Path.GetDirectoryName(exePath);
            process.StartInfo.RedirectStandardOutput = true;
            process.StartInfo.RedirectStandardError = true;
            process.StartInfo.UseShellExecute = false;
            process.StartInfo.CreateNoWindow = true;
            process.StartInfo.StandardOutputEncoding = Encoding.UTF8;
            process.StartInfo.StandardErrorEncoding = Encoding.UTF8;

            var outputBuilder = new StringBuilder();
            var errorBuilder = new StringBuilder();

            process.OutputDataReceived += (sender, args) =>
            {
                if (args.Data != null)
                    outputBuilder.AppendLine(args.Data);
            };
            process.ErrorDataReceived += (sender, args) =>
            {
                if (args.Data != null)
                    errorBuilder.AppendLine(args.Data);
            };

            bool isStarted = process.Start();
            if (!isStarted)
            {
                Name = "Échec du démarrage du processus.";
                return appStatuses;
            }

            process.BeginOutputReadLine();
            process.BeginErrorReadLine();

            await Task.Run(() => process.WaitForExit());

            string standardOutput = outputBuilder.ToString();
            string standardError = errorBuilder.ToString();

            if (process.ExitCode != 0)
            {
                Name = $"Erreur lors de la récupération des statuts : {standardError}";
                return appStatuses;
            }

            var lines = (standardError + standardOutput).Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);

            foreach (var line in lines)
            {
                if (line.Contains("|"))
                {
                    var parts = line.Split('|');
                    if (parts.Length >= 3)
                    {
                        string appName = parts[1].Trim();
                        string statusText = parts[2].Trim();
                        string version = "Inconnue";

                        AppStatus status;

                        if (statusText.Contains("not installed >> will install version"))
                        {
                            status = AppStatus.NotInstalled;
                            // Extraire la version
                            var match = Regex.Match(statusText, @"will install version ([^\s]+)");
                            if (match.Success)
                            {
                                version = match.Groups[1].Value;
                            }
                        }
                        else if (statusText.Contains("already up to date"))
                        {
                            status = AppStatus.Installed;
                            // Extraire la version
                            var match = Regex.Match(statusText, @"already up to date \(version ([^\)]+)\)");
                            if (match.Success)
                            {
                                version = match.Groups[1].Value;
                            }
                        }
                        else if (statusText.Contains("A newer version"))
                        {
                            status = AppStatus.NeedsUpdate;
                            // Extraire les versions
                            var match = Regex.Match(statusText, @"A newer version ([^\s]+) is available \(installed version ([^\)]+)\)");
                            if (match.Success)
                            {
                                string newVersion = match.Groups[1].Value;
                                string installedVersion = match.Groups[2].Value;
                                version = $"{installedVersion} (nouvelle : {newVersion})";
                            }
                        }
                        else
                        {
                            // Autres cas éventuels
                            status = AppStatus.NotInstalled;
                        }

                        appStatuses[appName] = (status, version);
                    }
                }
            }
            process = new Process();
            process.StartInfo.FileName = exePath;
            process.StartInfo.Arguments = "status";
            process.StartInfo.WorkingDirectory = System.IO.Path.GetDirectoryName(exePath);
            process.StartInfo.RedirectStandardOutput = true;
            process.StartInfo.RedirectStandardError = true;
            process.StartInfo.UseShellExecute = false;
            process.StartInfo.CreateNoWindow = true;
            process.StartInfo.StandardOutputEncoding = Encoding.UTF8;
            process.StartInfo.StandardErrorEncoding = Encoding.UTF8;

            isStarted = process.Start();
            if (!isStarted)
            {
                Name = "Failed to start the process.";
                return appStatuses;
            }

            standardOutput = process.StandardOutput.ReadToEnd();
            standardError = process.StandardError.ReadToEnd();
            process.WaitForExit();

            if (process.ExitCode != 0)
            {
                Name = $"Error getting statuses: {standardError}";
                return appStatuses;
            }

            lines = (standardError + standardOutput).Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);

            foreach (var line in lines)
            {
                if (line.Contains("|"))
                {
                    var parts = line.Split('|');
                    if (parts.Length >= 3)
                    {
                        string appName = parts[1].Trim();
                        string statusText = parts[2].Trim();
                        string version = parts.Length > 3 ? parts[3].Trim() : "Unknown";

                        AppStatus status;
                        if (statusText.Contains("already up to date"))
                        {
                            status = AppStatus.Installed;
                        }
                        else if (statusText.Contains("A newer version"))
                        {
                            status = AppStatus.NeedsUpdate;
                        }
                        else
                        {
                            status = AppStatus.NotInstalled;
                        }

                        foreach (string st in statusText.Split(" "))
                        {
                            
                            if (Regex.IsMatch(st, @"\d+"))
                            {
                                version = st;
                            }
                        }

                        appStatuses[appName] = (status, version);
                    }
                }
            }

            foreach (var appName in allApps)
            {
                if (!appStatuses.ContainsKey(appName))
                {
                    appStatuses[appName] = (AppStatus.NotInstalled, "Unknown");
                }
            }

            return appStatuses;
        }
        private async void InstallApp(string appName)
        {
            string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");

            Process process = new Process();
            process.StartInfo.FileName = exePath;
            process.StartInfo.Arguments = $"i {appName} --yes --verbose";
            process.StartInfo.WorkingDirectory = System.IO.Path.GetDirectoryName(exePath);
            process.StartInfo.RedirectStandardOutput = true;
            process.StartInfo.RedirectStandardError = true;
            process.StartInfo.RedirectStandardInput = true;
            process.StartInfo.UseShellExecute = false;
            process.StartInfo.CreateNoWindow = true;
            process.StartInfo.StandardOutputEncoding = Encoding.UTF8;
            process.StartInfo.StandardErrorEncoding = Encoding.UTF8;

            try
            {
                Name = $"Installation de {appName} en cours...";

                bool isStarted = process.Start();
                if (!isStarted)
                {
                    Name = $"Échec du démarrage du processus pour l'installation de {appName}.";
                    return;
                }

                // Envoi de "y" pour répondre à la demande de confirmation
                using (var writer = process.StandardInput)
                {
                    if (writer.BaseStream.CanWrite)
                    {
                        await writer.WriteLineAsync("y");
                    }
                }

                string standardOutput = await process.StandardOutput.ReadToEndAsync();
                string standardError = await process.StandardError.ReadToEndAsync();
                process.WaitForExit();

                if (process.ExitCode == 0)
                {
                    Name = $"Installation de {appName} terminée avec succès.";

                    var app = Applications.FirstOrDefault(a => a.Name == appName);
                    if (app != null)
                    {
                        app.Status = AppStatus.Installed;
                    }
                }
                else
                {
                    Name = $"Échec de l'installation de {appName}.\nErreurs : {standardError}";
                }
                RunNomadList(); // Automatically update app list after installation
            }
            catch (Exception ex)
            {
                Name = $"Exception lors de l'installation de {appName}: {ex.Message}";
            }
        }

        private void ExecuteSearch()
        {
            if (string.IsNullOrWhiteSpace(SearchText))
            {
                FilteredApplications.Clear();
                foreach (var app in Applications)
                {
                    FilteredApplications.Add(app);
                }
            }
            else
            {
                var filtered = Applications.Where(a => a.Name.Contains(SearchText, StringComparison.OrdinalIgnoreCase));
                FilteredApplications.Clear();
                foreach (var app in filtered)
                {
                    FilteredApplications.Add(app);
                }
            }
        }

        [ObservableProperty]
        private string name;

        partial void OnSearchTextChanged(string value)
        {
            ExecuteSearch();
        }
    }
}
