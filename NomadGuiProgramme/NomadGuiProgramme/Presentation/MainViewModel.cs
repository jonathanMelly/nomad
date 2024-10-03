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

    public partial class ApplicationItem : ObservableObject
    {
        public string Name { get; set; }

        private AppStatus status;
        public AppStatus Status
        {
            get => status;
            set => SetProperty(ref status, value);
        }

        public ICommand InstallCommand { get; }

        public ApplicationItem(string name, ICommand installCommand)
        {
            Name = name;
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
                    var status = appStatuses.ContainsKey(appName) ? appStatuses[appName] : AppStatus.NotInstalled;
                    var appItem = new ApplicationItem(appName, InstallAppCommand) { Status = status };
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
            process.StartInfo.Arguments = "list";
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

            var outputTask = process.StandardOutput.ReadToEndAsync();
            var errorTask = process.StandardError.ReadToEndAsync();

            await Task.WhenAll(outputTask, errorTask);

            process.WaitForExit();

            string standardOutput = outputTask.Result;
            string standardError = errorTask.Result;

            string allAppsStr = ((standardError + standardOutput).Split(":")).Last();
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

        private async Task<Dictionary<string, AppStatus>> GetAppStatuses(List<string> allApps)
        {
            var appStatuses = new Dictionary<string, AppStatus>();

            string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");

            Process process = new Process();
            process.StartInfo.FileName = exePath;
            process.StartInfo.Arguments = "status";
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
                return appStatuses;
            }

            var outputTask = process.StandardOutput.ReadToEndAsync();
            var errorTask = process.StandardError.ReadToEndAsync();

            await Task.WhenAll(outputTask, errorTask);

            process.WaitForExit();

            string standardOutput = outputTask.Result;
            string standardError = outputTask.Result;

            string output = standardError + standardOutput;

            var lines = output.Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);

            foreach (var line in lines)
            {
                if (line.Contains("|"))
                {
                    int startIndex = line.IndexOf("|") + 1;
                    int endIndex = line.IndexOf("|", startIndex);
                    if (endIndex > startIndex)
                    {
                        string appName = line.Substring(startIndex, endIndex - startIndex);
                        string statusText = line.Substring(endIndex + 1).Trim();

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

                        appStatuses[appName] = status;
                    }
                }
            }

            foreach (var appName in allApps)
            {
                if (!appStatuses.ContainsKey(appName))
                {
                    appStatuses[appName] = AppStatus.NotInstalled;
                }
            }

            return appStatuses;
        }

        private async void InstallApp(string appName)
        {
            string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");
            Debug.WriteLine($"Executable Path: {exePath}");

            Process process = new Process();
            process.StartInfo.FileName = exePath;
            process.StartInfo.Arguments = $"i {appName} --yes";
            process.StartInfo.WorkingDirectory = System.IO.Path.GetDirectoryName(exePath);
            process.StartInfo.RedirectStandardOutput = true;
            process.StartInfo.RedirectStandardError = true;
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

                var outputTask = process.StandardOutput.ReadToEndAsync();
                var errorTask = process.StandardError.ReadToEndAsync();

                await Task.WhenAll(outputTask, errorTask);
                process.WaitForExit();

                string standardOutput = outputTask.Result;
                string standardError = errorTask.Result;

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
