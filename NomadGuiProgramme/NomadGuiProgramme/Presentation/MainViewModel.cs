using System;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Input;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;

namespace NomadGuiProgramme.Presentation
{
    public enum AppStatus
    {
        NeedsUpdate,
        Installed,
        NotInstalled
    }

    public enum FilterOption
    {
        Tous,
        Installés,
        NonInstallés,
        MiseÀJourDisponible
    }

    public partial class ApplicationItem : ObservableObject
    {
        [ObservableProperty]
        private string name;

        [ObservableProperty]
        private string version;

        [ObservableProperty]
        private AppStatus status;

        public ICommand InstallCommand { get; }

        public ApplicationItem(string name, string version, ICommand installCommand)
        {
            Name = name;
            Version = version;
            InstallCommand = new RelayCommand(() => installCommand.Execute(Name));
        }

        public string ButtonContent => Status switch
        {
            AppStatus.NeedsUpdate => "Mettre à jour",
            AppStatus.NotInstalled => "Installer",
            AppStatus.Installed => "Réinstaller",
            _ => "Action",
        };

        public bool IsButtonEnabled => true;
    }

    public partial class MainViewModel : ObservableObject
    {
        public ICommand RunNomadListCommand { get; }
        public ICommand InstallAppCommand { get; }
        public ICommand SearchCommand { get; }

        public ObservableCollection<ApplicationItem> Applications { get; } = new ObservableCollection<ApplicationItem>();
        public ObservableCollection<ApplicationItem> FilteredApplications { get; } = new ObservableCollection<ApplicationItem>();

        public ObservableCollection<FilterOption> FilterOptions { get; } = new ObservableCollection<FilterOption>
        {
            FilterOption.Tous,
            FilterOption.Installés,
            FilterOption.NonInstallés,
            FilterOption.MiseÀJourDisponible
        };

        [ObservableProperty]
        private string searchText;

        [ObservableProperty]
        private FilterOption selectedFilter = FilterOption.Tous;

        [ObservableProperty]
        private string name;

        public MainViewModel()
        {
            RunNomadListCommand = new RelayCommand(RunNomadList);
            InstallAppCommand = new RelayCommand<string>(InstallApp);
            SearchCommand = new RelayCommand(ExecuteSearch);
            RunNomadList();
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
                    var (status, version) = appStatuses.ContainsKey(appName) ? appStatuses[appName] : (AppStatus.NotInstalled, "Unknown");
                    var appItem = new ApplicationItem(appName, version, InstallAppCommand)
                    {
                        Status = status
                    };
                    Applications.Add(appItem);
                }

                SortApplications();
                ExecuteSearch();
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

            var process = new Process
            {
                StartInfo = new ProcessStartInfo
                {
                    FileName = exePath,
                    Arguments = "list --verbose",
                    WorkingDirectory = System.IO.Path.GetDirectoryName(exePath),
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    StandardOutputEncoding = Encoding.UTF8,
                    StandardErrorEncoding = Encoding.UTF8
                }
            };

            bool isStarted = process.Start();
            if (!isStarted)
            {
                Name = "Failed to start the process.";
                return allApps;
            }

            string standardOutput = await process.StandardOutput.ReadToEndAsync();
            string standardError = await process.StandardError.ReadToEndAsync();
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

            foreach (var appName in allApps)
            {
                var process = new Process
                {
                    StartInfo = new ProcessStartInfo
                    {
                        FileName = exePath,
                        Arguments = $"status {appName}",
                        WorkingDirectory = System.IO.Path.GetDirectoryName(exePath),
                        RedirectStandardOutput = true,
                        RedirectStandardError = true,
                        UseShellExecute = false,
                        CreateNoWindow = true,
                        StandardOutputEncoding = Encoding.UTF8,
                        StandardErrorEncoding = Encoding.UTF8
                    }
                };

                bool isStarted = process.Start();
                if (!isStarted)
                {
                    Name = $"Failed to start the process for {appName}.";
                    continue;
                }

                string standardOutput = await process.StandardOutput.ReadToEndAsync();
                string standardError = await process.StandardError.ReadToEndAsync();
                process.WaitForExit();

                if (process.ExitCode != 0)
                {
                    Name = $"Error getting status for {appName}: {standardError}";
                    continue;
                }

                var lines = (standardError + standardOutput).Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);

                foreach (var line in lines)
                {
                    if (line.Contains("|"))
                    {
                        var parts = line.Split('|');
                        if (parts.Length >= 3)
                        {
                            string app = parts[1].Trim();
                            string statusText = parts[2].Trim();
                            string version = "Unknown";

                            AppStatus status;
                            if (statusText.Contains("already up to date"))
                            {
                                status = AppStatus.Installed;
                            }
                            else if (statusText.Contains("A newer version") || statusText.Contains("needs update"))
                            {
                                status = AppStatus.NeedsUpdate;
                            }
                            else if (statusText.Contains("not installed"))
                            {
                                status = AppStatus.NotInstalled;
                            }
                            else
                            {
                                status = AppStatus.NotInstalled;
                            }

                            // Extraction de la version
                            var versionMatch = System.Text.RegularExpressions.Regex.Match(statusText, @"\d+(\.\d+)+");
                            if (versionMatch.Success)
                            {
                                version = versionMatch.Value;
                            }
                            else if (statusText.Contains("will install version"))
                            {
                                var versionParts = statusText.Split("will install version");
                                if (versionParts.Length > 1)
                                {
                                    version = versionParts[1].Trim();
                                }
                            }

                            appStatuses[app] = (status, version);
                        }
                    }
                }

                // Si l'application n'a pas été ajoutée, on l'ajoute avec le statut par défaut
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
            var app = Applications.FirstOrDefault(a => a.Name == appName);

            if (app == null)
            {
                Name = $"Application {appName} introuvable.";
                return;
            }

            string action = app.Status == AppStatus.NeedsUpdate ? "upgrade" : "install";
            string arguments = $"{action} {appName} --yes --verbose";

            var process = new Process
            {
                StartInfo = new ProcessStartInfo
                {
                    FileName = exePath,
                    Arguments = arguments,
                    WorkingDirectory = System.IO.Path.GetDirectoryName(exePath),
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    RedirectStandardInput = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    StandardOutputEncoding = Encoding.UTF8,
                    StandardErrorEncoding = Encoding.UTF8
                }
            };

            try
            {
                Name = $"{(action == "upgrade" ? "Mise à jour" : "Installation")} de {appName} en cours...";

                bool isStarted = process.Start();
                if (!isStarted)
                {
                    Name = $"Échec du démarrage du processus pour {(action == "upgrade" ? "la mise à jour" : "l'installation")} de {appName}.";
                    return;
                }

                string standardOutput = await process.StandardOutput.ReadToEndAsync();
                string standardError = await process.StandardError.ReadToEndAsync();
                process.WaitForExit();

                if (process.ExitCode == 0)
                {
                    Name = $"{(action == "upgrade" ? "Mise à jour" : "Installation")} de {appName} terminée avec succès.";

                    app.Status = AppStatus.Installed;
                    app.Version = await GetAppVersion(appName);
                }
                else
                {
                    Name = $"Échec de {(action == "upgrade" ? "la mise à jour" : "l'installation")} de {appName}.\nErreurs : {standardError}";
                }

                RunNomadList(); // Met à jour la liste des applications
            }
            catch (Exception ex)
            {
                Name = $"Exception lors de {(action == "upgrade" ? "la mise à jour" : "l'installation")} de {appName}: {ex.Message}";
            }
        }

        private async Task<string> GetAppVersion(string appName)
        {
            string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");

            var process = new Process
            {
                StartInfo = new ProcessStartInfo
                {
                    FileName = exePath,
                    Arguments = $"status {appName}",
                    WorkingDirectory = System.IO.Path.GetDirectoryName(exePath),
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    StandardOutputEncoding = Encoding.UTF8,
                    StandardErrorEncoding = Encoding.UTF8
                }
            };

            bool isStarted = process.Start();
            if (!isStarted)
            {
                return "Unknown";
            }

            string standardOutput = await process.StandardOutput.ReadToEndAsync();
            string standardError = await process.StandardError.ReadToEndAsync();
            process.WaitForExit();

            if (process.ExitCode != 0)
            {
                return "Unknown";
            }

            var lines = (standardError + standardOutput).Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);

            foreach (var line in lines)
            {
                if (line.Contains("|"))
                {
                    var parts = line.Split('|');
                    if (parts.Length >= 3)
                    {
                        string statusText = parts[2].Trim();
                        var versionMatch = System.Text.RegularExpressions.Regex.Match(statusText, @"\d+(\.\d+)+");
                        if (versionMatch.Success)
                        {
                            return versionMatch.Value;
                        }
                    }
                }
            }

            return "Unknown";
        }

        partial void OnSelectedFilterChanged(FilterOption value)
        {
            ExecuteSearch();
        }

        partial void OnSearchTextChanged(string value)
        {
            ExecuteSearch();
        }

        private void ExecuteSearch()
        {
            var filtered = Applications.AsEnumerable();

            // Filtrage par texte de recherche
            if (!string.IsNullOrWhiteSpace(SearchText))
            {
                filtered = filtered.Where(a => a.Name.Contains(SearchText, StringComparison.OrdinalIgnoreCase));
            }

            // Filtrage par statut
            switch (SelectedFilter)
            {
                case FilterOption.Installés:
                    filtered = filtered.Where(a => a.Status == AppStatus.Installed);
                    break;
                case FilterOption.NonInstallés:
                    filtered = filtered.Where(a => a.Status == AppStatus.NotInstalled);
                    break;
                case FilterOption.MiseÀJourDisponible:
                    filtered = filtered.Where(a => a.Status == AppStatus.NeedsUpdate);
                    break;
                default:
                    break;
            }

            // Mise à jour de la collection filtrée
            FilteredApplications.Clear();
            foreach (var app in filtered)
            {
                FilteredApplications.Add(app);
            }
        }
    }
}
