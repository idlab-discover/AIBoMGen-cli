"use client";

import { LogOut, Moon, Settings, Sun, User } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { Avatar, AvatarFallback, AvatarImage } from "./ui/avatar";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import { Button } from "./ui/button";
import { useTheme } from "next-themes";
import { SidebarTrigger } from "./ui/sidebar";
import { useMsal } from "@azure/msal-react";
import { loginRequest } from "./configurations/msalConfig";
import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";

const Navbar = () => {
    const { theme, setTheme } = useTheme();
    const { instance, accounts } = useMsal();
    const t = useTranslations("Navbar");

    const [user, setUser] = useState<{ name: string; picture: string }>({
        name: t("guest"),
        picture: "https://ui-avatars.com/api/?name=g", // Default guest profile picture
    });

    // Update user state when accounts change
    useEffect(() => {
        if (accounts.length > 0) {
            const account = accounts[0];
            instance.setActiveAccount(account); // Always set active account when accounts change
            fetchProfilePicture(account);
        } else {
            setUser({
                name: t("guest"),
                picture: "https://ui-avatars.com/api/?name=g", // Reset to guest profile picture
            });
        }
    }, [accounts]);

    const fetchProfilePicture = async (account: any) => {
        try {
            // Request a token for Microsoft Graph
            const tokenResponse = await instance.acquireTokenSilent({
                scopes: ["User.Read"], // Only request Microsoft Graph scopes
                account,
            });

            const accessToken = tokenResponse.accessToken;

            // Fetch the profile picture from Microsoft Graph
            const response = await fetch("https://graph.microsoft.com/v1.0/me/photo/$value", {
                headers: {
                    Authorization: `Bearer ${accessToken}`,
                },
            });

            if (response.ok) {
                const blob = await response.blob();
                const pictureUrl = URL.createObjectURL(blob);

                setUser({
                    name: account.name || t("guest"),
                    picture: pictureUrl, // Set the profile picture URL
                });
            } else if (response.status === 401) {
                console.error("Unauthorized: Ensure the User.Read permission is granted.");
                setUser({
                    name: account.name || t("guest"),
                    picture: `https://ui-avatars.com/api/?name=${account.name}`, // Fallback to UI Avatars
                });
            } else {
                console.warn("No profile picture found for the user. Falling back to UI Avatars.");
                setUser({
                    name: account.name || t("guest"),
                    picture: `https://ui-avatars.com/api/?name=${account.name}`, // Fallback to UI Avatars
                });
            }
        } catch (error) {
            console.error("Error fetching profile picture:", error);
            setUser({
                name: account.name || t("guest"),
                picture: `https://ui-avatars.com/api/?name=${account.name}`, // Fallback to UI Avatars
            });
        }
    };

    // Handle login
    const handleLogin = async () => {
        try {
            // Step 1: Login with OpenID Connect scopes using redirect
            await instance.loginRedirect({
                scopes: ["openid", "profile", "offline_access", "User.Read"], // Only OpenID Connect scopes
                prompt: "consent", // Force consent screen for new users
            });
            // No further code needed; after redirect, MSAL will handle the response and update accounts
        } catch (error) {
            console.error("Login failed:", error);
        }
    };

    // Handle logout
    const handleLogout = async () => {
        try {
            await instance.logoutRedirect();
            instance.setActiveAccount(null); // Clear the active account after logout
            setUser({
                name: t("guest"),
                picture: "https://ui-avatars.com/api/?name=g&background=random", // Reset to guest profile picture
            });
        } catch (error) {
            console.error("Logout failed:", error);
        }
    };

    // Test the fetch job status endpoint
    const testJobStatus = async () => {
        try {
            const activeAccount = instance.getActiveAccount();
            if (!activeAccount) {
                console.error("No active account found. Please log in.");
                return;
            }

            // Request a token for your custom API
            const tokenResponse = await instance.acquireTokenSilent({
                scopes: ["api://a65becdd-c8b9-4a90-9c8a-9d9c526aa130/user_impersonation"], // Only request custom API scopes
                account: activeAccount,
            });

            const accessToken = tokenResponse.accessToken;

            // Call the job status endpoint
            const response = await fetch("http://localhost:8000/job_status/123e4567-e89b-12d3-a456-426614174000", {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${accessToken}`,
                },
            });

            if (response.ok) {
                const data = await response.json();
                console.log("Job status response:", data);
            } else {
                const errorData = await response.json();
                console.error("Error fetching job status:", errorData);
            }
        } catch (error) {
            console.error("Error testing job status endpoint:", error);
        }
    };



    return (
        <nav className="p-4 top-0 flex items-center justify-between sticky bg-background/50 backdrop-blur-3xl z-20">            {/* Left side of the navbar */}
            <SidebarTrigger />
            <div className="flex items-center gap-4 flex-shrink-0">
                <span className="text">{user.name}</span>
                {/* USER MENU */}
                <DropdownMenu>
                    <DropdownMenuTrigger>
                        <Avatar>
                            <AvatarImage src={user.picture} />
                            <AvatarFallback>{user.name.charAt(0)}</AvatarFallback>
                        </Avatar>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent sideOffset={4}>
                        <DropdownMenuLabel>{user.name}</DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem asChild>
                            <Link href="/profile">
                                <User className="h-[1.2rem] w-[1.2rem] mr-2" />
                                {t("profile")}
                            </Link>
                        </DropdownMenuItem>
                        <DropdownMenuItem asChild>
                            <Link href="/settings">
                                <Settings className="h-[1.2rem] w-[1.2rem] mr-2" />
                                {t("settings")}
                            </Link>
                        </DropdownMenuItem>
                        <DropdownMenuItem variant="destructive" onClick={handleLogout}>
                            <LogOut className="h-[1.2rem] w-[1.2rem] mr-2" />
                            {t("logout")}
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
                {/* LOGIN BUTTON */}
                {user.name === t("guest") && (
                    <Button onClick={handleLogin} variant="default">
                        {t("login")}
                    </Button>
                )}
                {/* THEME MENU */}
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="outline" size="icon">
                            <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
                            <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
                            <span className="sr-only">{t("toggleTheme")}</span>
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => setTheme("light")}>
                            {t("light")}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => setTheme("dark")}>
                            {t("dark")}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => setTheme("system")}>
                            {t("system")}
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
        </nav>
    );
};

export default Navbar;