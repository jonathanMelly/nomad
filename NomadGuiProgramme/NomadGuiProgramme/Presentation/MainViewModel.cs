using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Text;
using System.Xml.Linq;

public partial class MainViewModel : ObservableObject
{
    public ICommand RunNomadListCommand { get; }

    // Collection pour stocker les applications récupérées
    public ObservableCollection<string> Applications { get; } = new ObservableCollection<string>();

    public MainViewModel()
    {
        RunNomadListCommand = new RelayCommand(RunNomadList);
    }

    private async void RunNomadList()
    {
        string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");
        Debug.WriteLine($"Executable Path: {exePath}");

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

        try
        {
            bool isStarted = process.Start();
            if (!isStarted)
            {
                name = "Failed to start the process.";
                return;
            }

            var outputTask = process.StandardOutput.ReadToEndAsync();
            var errorTask = process.StandardError.ReadToEndAsync();

            await Task.WhenAll(outputTask, errorTask);

            process.WaitForExit();

            string standardOutput = outputTask.Result;
            string standardError = errorTask.Result;

            string allAppsStr = ((standardError + standardOutput).Split(":")).Last();
            string[] apps = allAppsStr.Split(",");

            Applications.Clear();  // Vider la collection avant d'ajouter les nouvelles apps
        foreach (var app in apps)
        {
            string trimmedApp = app.Trim();
            if (!string.IsNullOrEmpty(trimmedApp))
            {
                Applications.Add(trimmedApp);
            }
        }
        }
        catch (Exception ex)
        {
            Name = $"Exception: {ex.Message}";
        }
    }
    private async void InstallApp(string appName)
    {
        string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");
        Debug.WriteLine($"Executable Path: {exePath}");

        Process process = new Process();
        process.StartInfo.FileName = exePath;
        process.StartInfo.Arguments = $"i {appName}";
        process.StartInfo.WorkingDirectory = System.IO.Path.GetDirectoryName(exePath);
        process.StartInfo.RedirectStandardOutput = true;
        process.StartInfo.RedirectStandardError = true;
        process.StartInfo.RedirectStandardInput = true; // Rediriger l'entrée standard
        process.StartInfo.UseShellExecute = true;
        process.StartInfo.CreateNoWindow = false;

        process.StartInfo.StandardOutputEncoding = Encoding.UTF8;
        process.StartInfo.StandardErrorEncoding = Encoding.UTF8;

        try
        {
            bool isStarted = process.Start();
            if (!isStarted)
            {
                Name = "Failed to start the process.";
                return;
            }

            // Créer des tâches pour lire les sorties standard et les erreurs
            var outputTask = ReadOutputAsync(process.StandardOutput, process);
            var errorTask = ReadOutputAsync(process.StandardError, process);

            // Écrire les 'y' nécessaires dans l'entrée standard
            await Task.Delay(500); // Attendre un peu pour que le processus démarre et demande une confirmation
            await process.StandardInput.WriteLineAsync("y");

            // Attendre que le processus se termine
            await Task.WhenAll(outputTask, errorTask);
            process.WaitForExit();

            string standardOutput = outputTask.Result;
            string standardError = errorTask.Result;

            // Traiter les sorties
            Name = $"Installation de {appName} terminée.\nSortie : {standardOutput}\nErreurs : {standardError}";
        }
        catch (Exception ex)
        {
            Name = $"Exception lors de l'installation de {appName}: {ex.Message}";
        }
    }

    private async Task<string> ReadOutputAsync(StreamReader reader, Process process)
    {
        StringBuilder output = new StringBuilder();
        string line;
        while ((line = await reader.ReadLineAsync()) != null)
        {
            output.AppendLine(line);
            // Détecter les invites de confirmation si nécessaire et répondre
            if (line.Contains("Proceed [Y,n]") || line.Contains("Proceed [Y,n] (Enter=Yes) ?"))
            {
                await process.StandardInput.WriteLineAsync("y");
            }
        }
        return output.ToString();
    }

    [ObservableProperty]
    private string name;
}
